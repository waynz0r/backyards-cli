// Copyright © 2019 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package istio

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"emperror.dev/errors"
	"github.com/spf13/cobra"
	"istio.io/operator/pkg/object"
	appsv1 "k8s.io/api/apps/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/yaml"

	"github.com/banzaicloud/backyards-cli/cmd/backyards/static/istio_assets"
	"github.com/banzaicloud/backyards-cli/cmd/backyards/static/istio_operator"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/helm"
	"github.com/banzaicloud/backyards-cli/pkg/k8s"
	"github.com/banzaicloud/backyards-cli/pkg/util"
	"github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
)

const (
	IstioCRName         = "mesh"
	istioCRYamlFilename = "istio.yaml"
)

var (
	istioCRDs = []string{
		"istios.istio.banzaicloud.io",
		"remoteistios.istio.banzaicloud.io",
	}
)

type installCommand struct {
	cli cli.CLI
}

type InstallOptions struct {
	DumpResources bool

	istioCRFilename string
	releaseName     string
}

func NewInstallOptions() *InstallOptions {
	return &InstallOptions{}
}

func NewInstallCommand(cli cli.CLI, options *InstallOptions) *cobra.Command {
	c := &installCommand{
		cli: cli,
	}

	cmd := &cobra.Command{
		Use:   "install [flags]",
		Args:  cobra.NoArgs,
		Short: "Installs Istio utilizing Banzai Cloud's Istio-operator",
		Long: `Installs Istio utilizing Banzai Cloud's Istio-operator.

The command automatically applies the resources.
It can only dump the applicable resources with the '--dump-resources' option.

The manual mode is a two phase process as the operator needs custom CRDs to work.
The installer automatically detects whether the CRDs are installed or not, and behaves accordingly.`,
		Example: `  # Default install.
  backyards istio install

  # Install Istio into a non-default namespace.
  backyards istio install -n istio-custom-ns`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			return c.run(cli, options)
		},
	}

	cmd.Flags().StringVar(&options.releaseName, "release-name", "istio-operator", "Name of the release")
	cmd.Flags().StringVarP(&options.istioCRFilename, "istio-cr-file", "f", "", "Filename of a custom Istio CR yaml")

	cmd.Flags().BoolVarP(&options.DumpResources, "dump-resources", "d", options.DumpResources, "Dump resources to stdout instead of applying them")

	return cmd
}

func (c *installCommand) run(cli cli.CLI, options *InstallOptions) error {
	objects, err := getIstioOperatorObjects(options.releaseName)
	if err != nil {
		return err
	}
	objects.Sort(helm.InstallObjectOrder())

	istioCRObj, err := getIstioCR(options.istioCRFilename)
	if err != nil {
		return err
	}

	crds := make(object.K8sObjects, 0)
	objs := make(object.K8sObjects, 0)
	for _, obj := range objects {
		if obj.Kind == "CustomResourceDefinition" {
			crds = append(crds, obj)
		} else {
			objs = append(objs, obj)
		}
	}
	objs = append(objs, istioCRObj)

	if !options.DumpResources {
		err := c.applyResources(crds, objs)
		if err != nil {
			return errors.WrapIf(err, "could not apply resources")
		}
	} else {
		crdsExists, err := c.isCRDsExists(istioCRDs)
		if err != nil {
			return errors.WrapIf(err, "could not check whether CRD exists or not")
		}

		if !crdsExists {
			yaml, err := crds.YAMLManifest()
			if err != nil {
				return errors.WrapIf(err, "could not render YAML manifest")
			}
			fmt.Fprintln(os.Stderr, "The same command should be run after the CRDs are installed successfully to install the rest of the resources.")
			fmt.Fprint(cli.Out(), yaml)
		} else {
			yaml, err := objs.YAMLManifest()
			if err != nil {
				return errors.WrapIf(err, "could not render YAML manifest")
			}
			fmt.Fprint(cli.Out(), yaml)
		}
	}

	return nil
}

func (c *installCommand) applyResources(crds, objects object.K8sObjects) error {
	client, err := c.cli.GetK8sClient()
	if err != nil {
		return err
	}

	// apply CRDs first
	err = k8s.ApplyResources(client, crds)
	if err != nil {
		return errors.WrapIf(err, "could not apply k8s resources")
	}

	backoff := wait.Backoff{
		Duration: time.Second * 5,
		Factor:   1,
		Jitter:   0,
		Steps:    25,
	}

	err = k8s.WaitForResourcesConditions(client, k8s.NamesWithGVKFromK8sObjects(crds), backoff, k8s.CRDEstablishedConditionCheck)
	if err != nil {
		return err
	}

	// reinitialize client after CRDs creations
	client, err = c.cli.GetK8sClient()
	if err != nil {
		return err
	}

	// apply the rest of the resources
	err = k8s.ApplyResources(client, objects)
	if err != nil {
		return errors.WrapIf(err, "could not apply k8s resources")
	}

	err = k8s.WaitForResourcesConditions(client, k8s.NamesWithGVKFromK8sObjects(objects, "StatefulSet"), backoff, k8s.ExistsConditionCheck, k8s.ReadyReplicasConditionCheck)
	if err != nil {
		return err
	}
	err = k8s.WaitForResourcesConditions(client, c.getIstioDeploymentsToWaitFor(), backoff, k8s.ExistsConditionCheck, k8s.ReadyReplicasConditionCheck)
	if err != nil {
		return err
	}

	return nil
}

func (c *installCommand) getIstioDeploymentsToWaitFor() []k8s.NamespacedNameWithGVK {
	var istioCR v1beta1.Istio

	client, _ := c.cli.GetK8sClient()
	err := client.Get(context.Background(), types.NamespacedName{
		Name:      IstioCRName,
		Namespace: IstioNamespace,
	}, &istioCR)
	if err != nil {
		panic(err)
	}

	deploymentNames := make([]string, 0)

	if util.PointerToBool(istioCR.Spec.Citadel.Enabled) {
		deploymentNames = append(deploymentNames, "istio-citadel")
	}
	if util.PointerToBool(istioCR.Spec.SidecarInjector.Enabled) {
		deploymentNames = append(deploymentNames, "istio-sidecar-injector")
	}
	if util.PointerToBool(istioCR.Spec.Galley.Enabled) {
		deploymentNames = append(deploymentNames, "istio-galley")
	}
	if util.PointerToBool(istioCR.Spec.Pilot.Enabled) {
		deploymentNames = append(deploymentNames, "istio-pilot")
	}
	if util.PointerToBool(istioCR.Spec.Mixer.Enabled) {
		deploymentNames = append(deploymentNames, "istio-policy")
		deploymentNames = append(deploymentNames, "istio-telemetry")
	}
	if util.PointerToBool(istioCR.Spec.Gateways.Enabled) {
		if util.PointerToBool(istioCR.Spec.Gateways.IngressConfig.Enabled) {
			deploymentNames = append(deploymentNames, "istio-ingressgateway")
		}
		if util.PointerToBool(istioCR.Spec.Gateways.EgressConfig.Enabled) {
			deploymentNames = append(deploymentNames, "istio-egressgateway")
		}
	}

	deployments := make([]k8s.NamespacedNameWithGVK, len(deploymentNames))
	for i, name := range deploymentNames {
		deployments[i] = k8s.NamespacedNameWithGVK{
			NamespacedName: types.NamespacedName{
				Name:      name,
				Namespace: IstioNamespace,
			},
			GroupVersionKind: appsv1.SchemeGroupVersion.WithKind("Deployment"),
		}
	}

	return deployments
}

func getIstioOperatorObjects(releaseName string) (object.K8sObjects, error) {
	var values Values

	valuesYAML, err := helm.GetDefaultValues(istio_operator.Chart)
	if err != nil {
		return nil, errors.WrapIf(err, "could not get helm default values")
	}

	err = yaml.Unmarshal(valuesYAML, &values)
	if err != nil {
		return nil, errors.WrapIf(err, "could not unmarshal yaml values")
	}

	values.SetDefaults(releaseName)

	rawValues, err := yaml.Marshal(values)
	if err != nil {
		return nil, errors.WrapIf(err, "could not marshal yaml values")
	}

	objects, err := helm.Render(istio_operator.Chart, string(rawValues), helm.ReleaseOptions{
		Name:      "istio-operator",
		IsInstall: true,
		IsUpgrade: false,
		Namespace: IstioNamespace,
	}, "istio-operator")
	if err != nil {
		return nil, errors.WrapIf(err, "could not render helm manifest objects")
	}

	return objects, nil
}

func getIstioCR(filename string) (*object.K8sObject, error) {
	var err error
	var istioCRFile http.File
	if filename != "" {
		istioCRFile, err = os.Open(filename)
	} else {
		istioCRFile, err = istio_assets.Assets.Open(istioCRYamlFilename)
	}
	if err != nil {
		return nil, errors.WrapIf(err, "could not open Istio CR YAML")
	}
	defer istioCRFile.Close()

	yaml := new(bytes.Buffer)
	_, err = yaml.ReadFrom(istioCRFile)
	if err != nil {
		return nil, errors.WrapIf(err, "could not read Istio CR YAML")
	}

	obj, err := object.ParseYAMLToK8sObject(yaml.Bytes())
	if err != nil {
		return nil, errors.WrapIf(err, "could not parse Istio YAML to K8s object")
	}

	metadata := obj.UnstructuredObject().Object["metadata"].(map[string]interface{})
	metadata["namespace"] = IstioNamespace
	metadata["name"] = IstioCRName

	return obj, nil
}

func (c *installCommand) isCRDsExists(crdNames []string) (bool, error) {
	found := 0

	cl, err := c.cli.GetK8sClient()
	if err != nil {
		return false, err
	}

	var crd apiextensionsv1beta1.CustomResourceDefinition
	for _, crdName := range crdNames {
		err = cl.Get(context.Background(), types.NamespacedName{
			Name: crdName,
		}, &crd)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				err = nil
			}
			return false, errors.WrapIfWithDetails(err, "could not get CRD", "name", crdName)
		}
		found++
	}

	return found == len(crdNames), nil
}
