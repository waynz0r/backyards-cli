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
	"context"
	"fmt"
	"os"
	"time"

	"emperror.dev/errors"
	"github.com/spf13/cobra"
	"istio.io/operator/pkg/object"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/yaml"

	"github.com/banzaicloud/backyards-cli/cmd/backyards/static/istio_operator"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/helm"
	"github.com/banzaicloud/backyards-cli/pkg/k8s"
	"github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
)

const (
	pilotDockerImage = "banzaicloud/istio-pilot:1.2.2-bzc"
	mixerDockerImage = "banzaicloud/istio-mixer:1.2.2-bzc"
	istioCRName      = "mesh"
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

type installOptions struct {
	releaseName   string
	dumpResources bool
}

func newInstallCommand(cli cli.CLI) *cobra.Command {
	c := &installCommand{
		cli: cli,
	}
	options := &installOptions{}

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
	cmd.Flags().BoolVarP(&options.dumpResources, "dump-resources", "d", false, "Dump resources to stdout instead of applying them")

	return cmd
}

func (c *installCommand) run(cli cli.CLI, options *installOptions) error {
	objects, err := getIstioOperatorObjects(options.releaseName)
	if err != nil {
		return err
	}
	objects.Sort(helm.InstallObjectOrder())

	istioCRObj, err := getIstioCR()
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

	if !options.dumpResources {
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
	err = k8s.ApplyResources(client, crds, k8s.WaitForCRD(wait.Backoff{
		Duration: time.Second * 5,
		Factor:   1,
		Jitter:   0,
		Steps:    10,
	}))
	if err != nil {
		return errors.WrapIf(err, "could not apply k8s resources")
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

	return nil
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
		Namespace: istioNamespace,
	})
	if err != nil {
		return nil, errors.WrapIf(err, "could not render helm manifest objects")
	}

	return objects, nil
}

func getIstioCR() (*object.K8sObject, error) {
	istio := &v1beta1.Istio{
		ObjectMeta: metav1.ObjectMeta{
			Name:      istioCRName,
			Namespace: istioNamespace,
		},
	}

	v1beta1.SetDefaults(istio)

	istio.Spec.MTLS = true
	istio.Spec.ControlPlaneSecurityEnabled = true
	istio.Spec.SidecarInjector.RewriteAppHTTPProbe = true
	istio.Spec.Version = "1.2"
	istio.Spec.ImagePullPolicy = corev1.PullAlways
	istio.Spec.Gateways.IngressConfig.MaxReplicas = 1
	istio.Spec.Gateways.EgressConfig.MaxReplicas = 1
	istio.Spec.Pilot = v1beta1.PilotConfiguration{
		Image:       pilotDockerImage,
		MaxReplicas: 1,
	}
	istio.Spec.Mixer = v1beta1.MixerConfiguration{
		Image:       mixerDockerImage,
		MaxReplicas: 1,
	}

	y, err := yaml.Marshal(istio)
	if err != nil {
		return nil, errors.WrapIf(err, "could not marshal Istio resource to YAML")
	}

	yaml := fmt.Sprintf("apiVersion: %s\nkind: %s\n%s", v1beta1.SchemeGroupVersion.String(), "Istio", string(y))

	obj, err := object.ParseYAMLToK8sObject([]byte(yaml))
	if err != nil {
		return nil, errors.WrapIf(err, "could not parse Istio YAML to K8s object")
	}

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
