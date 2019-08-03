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

package demoapp

import (
	"context"
	"fmt"
	"os"

	"emperror.dev/errors"
	"github.com/spf13/cobra"
	"istio.io/operator/pkg/object"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/banzaicloud/backyards-cli/cmd/backyards/static"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/helm"
	"github.com/banzaicloud/backyards-cli/pkg/k8s"
)

const (
	istioNotFoundErrorTemplate = `Unable to install Backyards: %s

An existing Istio installation is required. You can install it with:

backyards istio install | kubectl apply -f -
`
)

var (
	sidecarPodLabels = map[string]string{
		"app": "istio-sidecar-injector",
	}
)

type installCommand struct {
	cli cli.CLI
}

type installOptions struct {
	namespace      string
	istioNamespace string

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
		Short: "Install demo application",
		Long: `Installs demo application.

The command provides the resources that can be applied manually or
it can apply the resources automatically with --apply-resources option.`,
		Example: `  # Default install.
  backyards demoapp install | kubectl apply -f -

  # Install Backyards into a non-default namespace.
  backyards demoapp install -n backyards-system | kubectl apply -f -`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			return c.run(cli, options)
		},
	}

	cmd.Flags().StringVar(&options.namespace, "demo-namespace", "backyards-demo", "Namespace for demo application")
	cmd.Flags().StringVar(&options.istioNamespace, "istio-namespace", "istio-system", "Namespace of Istio sidecar injector")
	cmd.Flags().BoolVarP(&options.dumpResources, "dump-resources", "d", false, "Dump resources to stdout instead of applying them")

	return cmd
}

func (c *installCommand) run(cli cli.CLI, options *installOptions) error {
	err := c.validate(options.istioNamespace)
	if err != nil {
		fmt.Fprintf(os.Stderr, istioNotFoundErrorTemplate, err)
		return nil
	}

	objects, err := getMeshdemoObjects(options.namespace)
	if err != nil {
		return err
	}
	objects.Sort(helm.InstallObjectOrder())

	if !options.dumpResources {
		client, err := cli.GetK8sClient()
		if err != nil {
			return err
		}

		err = k8s.ApplyResources(client, objects)
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

func getMeshdemoObjects(namespace string) (object.K8sObjects, error) {
	var values Values

	valuesYAML, err := helm.GetDefaultValues(static.MeshdemoChart)
	if err != nil {
		return nil, errors.WrapIf(err, "could not get helm default values")
	}

	err = yaml.Unmarshal(valuesYAML, &values)
	if err != nil {
		return nil, errors.WrapIf(err, "could not unmarshal yaml values")
	}

	values.UseNamespaceResource = true

	rawValues, err := yaml.Marshal(values)
	if err != nil {
		return nil, errors.WrapIf(err, "could not marshal yaml values")
	}

	objects, err := helm.Render(static.MeshdemoChart, string(rawValues), helm.ReleaseOptions{
		Name:      "meshdemo",
		IsInstall: true,
		IsUpgrade: false,
		Namespace: namespace,
	})
	if err != nil {
		return nil, errors.WrapIf(err, "could not render helm manifest objects")
	}

	return objects, nil
}

func (c *installCommand) validate(istioNamespace string) error {
	cl, err := c.cli.GetK8sClient()
	if err != nil {
		return errors.WrapIf(err, "could not get k8s client")
	}
	var pods v1.PodList
	err = cl.List(context.Background(), &pods, client.InNamespace(istioNamespace), client.MatchingLabels(sidecarPodLabels))
	if err != nil {
		return errors.WrapIf(err, "could not list pods")
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase == v1.PodRunning {
			return nil
		}
	}

	if len(pods.Items) > 0 {
		errors.Errorf("Istio sidecar injector not healthy yet in '%s'", istioNamespace)
	}

	return errors.Errorf("could not find Istio sidecar injector in '%s'", istioNamespace)
}
