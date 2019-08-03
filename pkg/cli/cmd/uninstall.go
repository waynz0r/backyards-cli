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

	"emperror.dev/errors"
	"github.com/spf13/cobra"

	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/helm"
	"github.com/banzaicloud/backyards-cli/pkg/k8s"
)

type uninstallCommand struct{}

type uninstallOptions struct {
	releaseName    string
	istioNamespace string
	dumpResources  bool
}

func newUninstallCommand(cli cli.CLI) *cobra.Command {
	c := &uninstallCommand{}
	options := &uninstallOptions{}

	cmd := &cobra.Command{
		Use:   "uninstall [flags]",
		Args:  cobra.NoArgs,
		Short: "Uninstalls Backyards",
		Long: `Uninstall Backyards

The command provides the resources that can be deleted manually or
it can delete the resources automatically with --delete-resources option.`,
		Example: `  # Default uninstall
  backyards uninstall | kubectl delete -f -

  # Uninstall Backyards from a non-default namespace
  backyards uninstall -n backyards-system | kubectl delete -f -`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			return c.run(cli, options)
		},
	}

	cmd.Flags().StringVar(&options.releaseName, "release-name", "backyards", "Name of the release")
	cmd.Flags().StringVar(&options.istioNamespace, "istio-namespace", "istio-system", "Namespace of Istio sidecar injector")
	cmd.Flags().BoolVarP(&options.dumpResources, "dump-resources", "d", false, "Dump resources to stdout instead of applying them")

	return cmd
}

func (c *uninstallCommand) run(cli cli.CLI, options *uninstallOptions) error {
	objects, err := getBackyardsObjects(options.releaseName, options.istioNamespace)
	if err != nil {
		return err
	}
	objects.Sort(helm.UninstallObjectOrder())

	if !options.dumpResources {
		client, err := cli.GetK8sClient()
		if err != nil {
			return err
		}

		err = k8s.DeleteResources(client, objects)
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
