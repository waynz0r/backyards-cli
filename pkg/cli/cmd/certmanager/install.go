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

package certmanager

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"emperror.dev/errors"
	"github.com/MakeNowJust/heredoc"
	"github.com/banzaicloud/backyards-cli/cmd/backyards/static/certmanager"
	"github.com/banzaicloud/backyards-cli/cmd/backyards/static/certmanagercainjector"
	"github.com/banzaicloud/backyards-cli/cmd/backyards/static/certmanagercrds"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/helm"
	"github.com/banzaicloud/backyards-cli/pkg/k8s"
	"github.com/spf13/cobra"
	"istio.io/operator/pkg/object"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

type installCommand struct {
	cli cli.CLI
}

type installOptions struct {
	DumpResources bool
}

func NewInstallOptions() *installOptions {
	return &installOptions{}
}

func NewInstallCommand(cli cli.CLI, options *installOptions) *cobra.Command {
	c := &installCommand{
		cli: cli,
	}

	cmd := &cobra.Command{
		Use:   "install [flags]",
		Args:  cobra.NoArgs,
		Short: "Install cert-manager",
		Long: `Installs cert-manager.

The command automatically applies the resources.
It can only dump the applicable resources with the '--dump-resources' option.`,
		Example: `  # Install to the cert-manager namespace. This command will fail if cert-manager is already installed from a different source.
  backyards cert-manager install
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			return c.run(cli, options)
		},
	}

	cmd.Flags().BoolVarP(&options.DumpResources, "dump-resources", "d", options.DumpResources, "Dump resources to stdout instead of applying them")

	return cmd
}

func (c *installCommand) run(cli cli.CLI, options *installOptions) error {
	err := c.validate(CertManagerNamespace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cert-manager validation failed: %s", err)
		return nil
	}

	objects, err := getCertManagerObjects(CertManagerNamespace)
	if err != nil {
		return err
	}
	objects.Sort(helm.InstallObjectOrder())

	if !options.DumpResources {
		client, err := cli.GetK8sClient()
		if err != nil {
			return err
		}

		err = k8s.ApplyResources(client, objects)
		if err != nil {
			return err
		}

		err = k8s.WaitForResourcesConditions(client, k8s.NamesWithGVKFromK8sObjects(objects), wait.Backoff{
			Duration: time.Second * 5,
			Factor:   1,
			Jitter:   0,
			Steps:    24,
		}, k8s.ExistsConditionCheck, k8s.ReadyReplicasConditionCheck)
		if err != nil {
			return err
		}
	} else {
		yaml, err := objects.YAMLManifest()
		if err != nil {
			return err
		}
		fmt.Fprintf(cli.Out(), yaml)
	}

	return nil
}

func (c *installCommand) validate(namespace string) error {
	if exists, err := c.crdExists("certificates.certmanager.k8s.io"); err != nil {
		return errors.WrapIf(err, "failed to check cert-manager certificate CRD")
	} else if exists {
		config, err := c.cli.GetK8sConfig()
		if err != nil {
			return errors.WrapIf(err, "failed to get k8s config to validate cert-manager")
		}
		client, err := kubernetes.NewForConfig(config)
		if err != nil {
			return errors.WrapIf(err, "failed to get k8s client to validate cert-manager")
		}
		targetNamespace, err := client.CoreV1().Namespaces().Get(namespace, v1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return errors.New("cert-manager already installed and not managed by us, " +
					"please remove previous cert-manager to continue")
			}
			return errors.WrapIf(err, "failed to get target namespace for cert-manager")
		}
		if manager, ok := targetNamespace.Labels["app.kubernetes.io/managed-by"]; ok && manager == "backyards-cli" {
			return nil
		} else if ok {
			return errors.Errorf("cert-manager already installed but the owner is unknown: %s; "+
				"please remove previous cert-manager to continue", manager)
		}
		return errors.New("cert-manager already installed but the owner is unknown; " +
			"please remove previous cert-manager to continue")
	}
	return nil
}

func getCertManagerNamespace(namespace string) (object.K8sObjects, error) {
	manifest := fmt.Sprintf(heredoc.Doc(`
		apiVersion: v1
		kind: Namespace
		metadata:
		  labels:
		    certmanager.k8s.io/disable-validation: "true"
		    app: cert-manager
		    app.kubernetes.io/name: cert-manager
		    app.kubernetes.io/managed-by: backyards-cli
		    app.kubernetes.io/instance: cert-manager
		    app.kubernetes.io/part-of: backyards
		  name: %s
	`), namespace)
	return object.ParseK8sObjectsFromYAMLManifest(manifest)
}

func getCertManagerCRDs() (object.K8sObjects, error) {
	crds, err := certmanagercrds.CRDs.Open("crds.yaml")
	if err != nil {
		errors.WrapIf(err, "failed to open certmanager crds file")
	}
	defer crds.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(crds)

	return object.ParseK8sObjectsFromYAMLManifest(buf.String())
}

func getCertManagerObjects(namespace string) (object.K8sObjects, error) {
	valuesYAML, err := helm.GetDefaultValues(certmanager.Chart)
	if err != nil {
		return nil, errors.WrapIf(err, "could not get helm default values")
	}

	objects, err := helm.Render(certmanager.Chart, string(valuesYAML), helm.ReleaseOptions{
		Name:      certManagerReleaseName,
		IsInstall: true,
		IsUpgrade: false,
		Namespace: namespace,
	}, "cert-manager")
	if err != nil {
		return nil, errors.WrapIf(err, "could not render helm manifest objects")
	}

	caInjectorValuesYAML, err := helm.GetDefaultValues(certmanagercainjector.Chart)
	if err != nil {
		return nil, errors.WrapIf(err, "could not get helm default values")
	}

	cainjectorObjects, err := helm.Render(certmanagercainjector.Chart, string(caInjectorValuesYAML), helm.ReleaseOptions{
		Name:      certManagerReleaseName,
		IsInstall: true,
		IsUpgrade: false,
		Namespace: namespace,
	}, "cainjector")
	if err != nil {
		return nil, errors.WrapIf(err, "could not render helm manifest objects")
	}

	namespaceObj, err := getCertManagerNamespace(namespace)
	if err != nil {
		return nil, errors.WrapIf(err, "could not render cert-manager namespace object")
	}

	crdObjects, err := getCertManagerCRDs()
	if err != nil {
		return nil, errors.WrapIf(err, "could not render cert-manager crd objects")
	}

	return append(crdObjects, append(namespaceObj, append(cainjectorObjects, objects...)...)...), nil
}

func (c *installCommand) crdExists(crdName string) (bool, error) {
	cl, err := c.cli.GetK8sClient()
	if err != nil {
		return false, err
	}

	var crd apiextensions.CustomResourceDefinition

	err = cl.Get(context.Background(), types.NamespacedName{
		Name: crdName,
	}, &crd)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, errors.WrapIfWithDetails(err, "could not get CRD", "name", crdName)
	}

	return true, nil
}
