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
	"fmt"
	"time"

	"emperror.dev/errors"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/canary"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/certmanager"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/demoapp"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/istio"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/helm"
	"github.com/banzaicloud/backyards-cli/pkg/k8s"
)

type uninstallCommand struct{}

type UninstallOptions struct {
	releaseName    string
	istioNamespace string
	dumpResources  bool

	uninstallCanary      bool
	uninstallDemoapp     bool
	uninstallIstio       bool
	uninstallCertManager bool
	uninstallEverything  bool
}

func NewUninstallCommand(cli cli.CLI) *cobra.Command {
	c := &uninstallCommand{}
	options := &UninstallOptions{}

	cmd := &cobra.Command{
		Use:   "uninstall [flags]",
		Args:  cobra.NoArgs,
		Short: "Uninstall Backyards",
		Long: `Uninstall Backyards

The command automatically removes the resources.
It can only dump the removable resources with the '--dump-resources' option.`,
		Example: `  # Default uninstall
  backyards uninstall

  # Uninstall Backyards from a non-default namespace
  backyards uninstall -n backyards-system`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			err := c.run(cli, options)
			if err != nil {
				return err
			}

			return c.runSubcommands(cli, options)
		},
	}

	cmd.Flags().StringVar(&options.releaseName, "release-name", "backyards", "Name of the release")
	cmd.Flags().StringVar(&options.istioNamespace, "istio-namespace", "istio-system", "Namespace of Istio sidecar injector")
	cmd.Flags().BoolVarP(&options.dumpResources, "dump-resources", "d", false, "Dump resources to stdout instead of applying them")

	cmd.Flags().BoolVar(&options.uninstallCanary, "uninstall-canary", false, "Uninstall Canary feature as well")
	cmd.Flags().BoolVar(&options.uninstallDemoapp, "uninstall-demoapp", false, "Uninstall Demo application as well")
	cmd.Flags().BoolVar(&options.uninstallIstio, "uninstall-istio", false, "Uninstall Istio mesh as well")
	cmd.Flags().BoolVar(&options.uninstallCertManager, "uninstall-cert-manager", false, "Uninstall cert-manager as well")
	cmd.Flags().BoolVarP(&options.uninstallEverything, "uninstall-everything", "a", false, "Uninstall every component at once")

	return cmd
}

func (c *uninstallCommand) run(cli cli.CLI, options *UninstallOptions) error {
	objects, err := getBackyardsObjects(options.releaseName, options.istioNamespace, nil)
	if err != nil {
		return err
	}
	objects.Sort(helm.UninstallObjectOrder())

	if !options.dumpResources {
		client, err := cli.GetK8sClient()
		if err != nil {
			return err
		}

		err = k8s.DeleteResources(client, objects, k8s.WaitForResourceConditions(wait.Backoff{
			Duration: time.Second * 5,
			Factor:   1,
			Jitter:   0,
			Steps:    24,
		}, k8s.NonExistsConditionCheck))
		if err != nil {
			return errors.WrapIf(err, "could not delete k8s resources")
		}
		return nil
	}

	yaml, err := objects.YAMLManifest()
	if err != nil {
		return errors.WrapIf(err, "could not render YAML manifest")
	}
	fmt.Fprint(cli.Out(), yaml)

	return nil
}

func (c *uninstallCommand) runSubcommands(cli cli.CLI, options *UninstallOptions) error {
	var err error
	var scmd *cobra.Command

	if options.uninstallDemoapp || options.uninstallEverything {
		scmdOptions := demoapp.NewUninstallOptions()
		if options.dumpResources {
			scmdOptions.DumpResources = true
		}
		scmd = demoapp.NewUninstallCommand(cli, scmdOptions)
		err = scmd.RunE(scmd, nil)
		if err != nil {
			return errors.WrapIf(err, "error during demo application uninstall")
		}
	}

	if options.uninstallCanary || options.uninstallEverything {
		scmdOptions := canary.NewUninstallOptions()
		if options.dumpResources {
			scmdOptions.DumpResources = true
		}
		scmd = canary.NewUninstallCommand(cli, scmdOptions)
		err = scmd.RunE(scmd, nil)
		if err != nil {
			return errors.WrapIf(err, "error during Canary feature uninstall")
		}
	}

	if options.uninstallCertManager || options.uninstallEverything {
		scmdOptions := certmanager.NewUninstallOptions()
		if options.dumpResources {
			scmdOptions.DumpResources = true
		}
		scmd = certmanager.NewUninstallCommand(cli, scmdOptions)
		err = scmd.RunE(scmd, nil)
		if err != nil {
			return errors.WrapIf(err, "error during cert-manager uninstall")
		}
	}

	if options.uninstallIstio || options.uninstallEverything {
		scmdOptions := istio.NewUninstallOptions()
		if options.dumpResources {
			scmdOptions.DumpResources = true
		}
		scmd = istio.NewUninstallCommand(cli, scmdOptions)
		err = scmd.RunE(scmd, nil)
		if err != nil {
			return errors.WrapIf(err, "error during Istio mesh uninstall")
		}
	}

	return nil
}
