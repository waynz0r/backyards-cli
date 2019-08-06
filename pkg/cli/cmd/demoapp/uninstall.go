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
	"fmt"
	"time"

	"emperror.dev/errors"
	"github.com/spf13/cobra"
	"istio.io/operator/pkg/object"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/helm"
	"github.com/banzaicloud/backyards-cli/pkg/k8s"
)

type uninstallCommand struct {
	cli cli.CLI
}

type uninstallOptions struct {
	namespace string

	dumpResources bool
}

func newUninstallCommand(cli cli.CLI) *cobra.Command {
	c := &uninstallCommand{
		cli: cli,
	}
	options := &uninstallOptions{}

	cmd := &cobra.Command{
		Use:   "uninstall [flags]",
		Args:  cobra.NoArgs,
		Short: "Output or delete Kubernetes resources to uninstall demo application",
		Long: `Output or delete Kubernetes resources to uninstall demo application.

The command automatically removes the resources.
It can only dump the removable resources with the '--dump-resources' option.`,
		Example: `  # Default uninstall.
  backyards canary uninstall

  # Uninstall Canary feature from a non-default namespace.
  backyards canary uninstall install -n custom-istio-ns`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			return c.run(cli, options)
		},
	}

	cmd.Flags().StringVar(&options.namespace, "demo-namespace", "backyards-demo", "Namespace for demo application")
	cmd.Flags().BoolVarP(&options.dumpResources, "dump-resources", "d", false, "Dump resources to stdout instead of applying them")

	return cmd
}

func (c *uninstallCommand) run(cli cli.CLI, options *uninstallOptions) error {
	objects, err := getMeshdemoObjects(options.namespace)
	if err != nil {
		return err
	}
	objects.Sort(helm.UninstallObjectOrder())

	if !options.dumpResources {
		err := c.deleteResources(objects)
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

func (c *uninstallCommand) deleteResources(objects object.K8sObjects) error {
	client, err := c.cli.GetK8sClient()
	if err != nil {
		return err
	}

	err = k8s.DeleteResources(client, objects, k8s.WaitForFinalizers(wait.Backoff{
		Duration: time.Second * 5,
		Factor:   1,
		Jitter:   0,
		Steps:    10,
	}))
	if err != nil {
		return errors.WrapIf(err, "could not delete k8s resources")
	}

	return nil
}
