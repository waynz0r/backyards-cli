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

package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"emperror.dev/errors"
	"github.com/spf13/cobra"
	"istio.io/operator/pkg/object"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/banzaicloud/backyards-cli/cmd/backyards/static/backyards"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/cli/cmd/canary"
	"github.com/banzaicloud/backyards-cli/pkg/cli/cmd/demoapp"
	"github.com/banzaicloud/backyards-cli/pkg/cli/cmd/istio"
	"github.com/banzaicloud/backyards-cli/pkg/helm"
	"github.com/banzaicloud/backyards-cli/pkg/k8s"
)

const (
	istioNotFoundErrorTemplate = `Unable to install Backyards: %s

An existing Istio installation is required. You can install it with:

backyards istio install
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
	releaseName    string
	istioNamespace string
	dumpResources  bool

	installCanary     bool
	installDemoapp    bool
	installIstio      bool
	installEverything bool
}

func newInstallCommand(cli cli.CLI) *cobra.Command {
	c := &installCommand{
		cli: cli,
	}
	options := &installOptions{}

	cmd := &cobra.Command{
		Use:   "install [flags]",
		Args:  cobra.NoArgs,
		Short: "Install Backyards",
		Long: `Installs Backyards.

The command automatically applies the resources.
It can only dump the applicable resources with the '--dump-resources' option.

The command can install every component at once with the '--install-everything' option.`,
		Example: `  # Default install.
  backyards install

  # Install Backyards into a non-default namespace.
  backyards install -n backyards-system`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			err := c.runSubcommands(cli, options)
			if err != nil {
				return err
			}

			return c.run(cli, options)
		},
	}

	cmd.Flags().StringVar(&options.releaseName, "release-name", "backyards", "Name of the release")
	cmd.Flags().StringVar(&options.istioNamespace, "istio-namespace", "istio-system", "Namespace of Istio sidecar injector")

	cmd.Flags().BoolVar(&options.installCanary, "install-canary", false, "Install Canary feature as well")
	cmd.Flags().BoolVar(&options.installDemoapp, "install-demoapp", false, "Install Demo application as well")
	cmd.Flags().BoolVar(&options.installIstio, "install-istio", false, "Install Istio mesh as well")
	cmd.Flags().BoolVarP(&options.installEverything, "install-everything", "a", false, "Install every component at once")

	cmd.Flags().BoolVarP(&options.dumpResources, "dump-resources", "d", false, "Dump resources to stdout instead of applying them")

	return cmd
}

func (c *installCommand) run(cli cli.CLI, options *installOptions) error {
	err := c.validate(options.istioNamespace)
	if err != nil {
		fmt.Fprintf(os.Stderr, istioNotFoundErrorTemplate, err)
		return nil
	}

	objects, err := getBackyardsObjects(options.releaseName, options.istioNamespace)
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

func getBackyardsObjects(releaseName, istioNamespace string) (object.K8sObjects, error) {
	var values Values

	valuesYAML, err := helm.GetDefaultValues(backyards.Chart)
	if err != nil {
		return nil, errors.WrapIf(err, "could not get helm default values")
	}

	err = yaml.Unmarshal(valuesYAML, &values)
	if err != nil {
		return nil, errors.WrapIf(err, "could not unmarshal yaml values")
	}

	values.SetDefaults(releaseName, istioNamespace)

	rawValues, err := yaml.Marshal(values)
	if err != nil {
		return nil, errors.WrapIf(err, "could not marshal yaml values")
	}

	objects, err := helm.Render(backyards.Chart, string(rawValues), helm.ReleaseOptions{
		Name:      "backyards",
		IsInstall: true,
		IsUpgrade: false,
		Namespace: backyardsNamespace,
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

func (c *installCommand) runSubcommands(cli cli.CLI, options *installOptions) error {
	var err error
	var scmd *cobra.Command

	if options.installIstio || options.installEverything {
		scmdOptions := istio.NewInstallOptions()
		if options.dumpResources {
			scmdOptions.DumpResources = true
		}
		scmd = istio.NewInstallCommand(cli, scmdOptions)
		err = scmd.RunE(scmd, nil)
		if err != nil {
			return errors.WrapIf(err, "error during Istio mesh install")
		}
	}

	if options.installCanary || options.installEverything {
		scmdOptions := canary.NewInstallOptions()
		if options.dumpResources {
			scmdOptions.DumpResources = true
		}
		scmd = canary.NewInstallCommand(cli, scmdOptions)
		err = scmd.RunE(scmd, nil)
		if err != nil {
			return errors.WrapIf(err, "error during Canary feature install")
		}
	}

	if options.installDemoapp || options.installEverything {
		scmdOptions := demoapp.NewInstallOptions()
		if options.dumpResources {
			scmdOptions.DumpResources = true
		}
		scmd = demoapp.NewInstallCommand(cli, scmdOptions)
		err = scmd.RunE(scmd, nil)
		if err != nil {
			return errors.WrapIf(err, "error during demo application install")
		}
	}

	return nil
}
