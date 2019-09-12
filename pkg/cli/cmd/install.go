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
	"github.com/banzaicloud/backyards-cli/pkg/cli/cmd/certmanager"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"
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
	requirementNotFoundErrorTemplate = "Unable to install Backyards: %s\n"
	defaultReleaseName               = "backyards"
)

var (
	sidecarPodLabels = map[string]string{
		"app": "istio-sidecar-injector",
	}
	certManagerPodLabels = map[string]string{
		"app": "cert-manager",
	}
)

type installCommand struct {
	cli cli.CLI
}

type installOptions struct {
	releaseName    string
	istioNamespace string
	dumpResources  bool

	installCanary      bool
	installDemoapp     bool
	installIstio       bool
	installCertManager bool
	disableCertManager bool
	disableAuditSink   bool
	installEverything  bool
	runDemo            bool
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
			var err error

			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			err = c.runSubcommands(cli, options)
			if err != nil {
				return err
			}

			err = c.run(cli, options)
			if err != nil {
				return err
			}

			err = c.runDemo(cli, options)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&options.releaseName, "release-name", defaultReleaseName, "Name of the release")
	cmd.Flags().StringVar(&options.istioNamespace, "istio-namespace", istio.DefaultNamespace, "Namespace of Istio sidecar injector")

	cmd.Flags().BoolVar(&options.installCanary, "install-canary", options.installCanary, "Install Canary feature as well")
	cmd.Flags().BoolVar(&options.installDemoapp, "install-demoapp", options.installDemoapp, "Install Demo application as well")
	cmd.Flags().BoolVar(&options.installIstio, "install-istio", options.installIstio, "Install Istio mesh as well")
	cmd.Flags().BoolVar(&options.installCertManager, "install-cert-manager", options.installIstio, "Install cert-manager as well")
	cmd.Flags().BoolVarP(&options.installEverything, "install-everything", "a", options.installEverything, "Install every component at once")

	cmd.Flags().BoolVar(&options.runDemo, "run-demo", options.runDemo, "Send load to demo application and opens up dashboard")
	cmd.Flags().BoolVar(&options.disableCertManager, "disable-cert-manager", options.disableCertManager, "Disable dependency on cert-manager and on it's resources")
	cmd.Flags().BoolVar(&options.disableAuditSink, "disable-auditsink", options.disableAuditSink, "Disable deploying the auditsink service and sending audit logs over http")

	cmd.Flags().BoolVarP(&options.dumpResources, "dump-resources", "d", options.dumpResources, "Dump resources to stdout instead of applying them")

	return cmd
}

func (c *installCommand) run(cli cli.CLI, options *installOptions) error {
	err := c.validate(options)
	if err != nil {
		errors := multierr.Errors(err)
		var errorItems string
		for _, e := range errors {
			errorItems += "\n - " + e.Error()
		}
		fmt.Fprintf(os.Stderr, requirementNotFoundErrorTemplate, errorItems)
		return nil
	}

	objects, err := getBackyardsObjects(options.releaseName, options.istioNamespace, func (values *Values) {
		values.CertManager.Enabled = !options.disableCertManager
		values.AuditSink.Enabled = !options.disableAuditSink
	})

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

func getBackyardsObjects(releaseName, istioNamespace string, valueOverrideFunc func (values *Values)) (object.K8sObjects, error) {
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

	if valueOverrideFunc != nil {
		valueOverrideFunc(&values)
	}

	rawValues, err := yaml.Marshal(values)
	if err != nil {
		return nil, errors.WrapIf(err, "could not marshal yaml values")
	}

	objects, err := helm.Render(backyards.Chart, string(rawValues), helm.ReleaseOptions{
		Name:      "backyards",
		IsInstall: true,
		IsUpgrade: false,
		Namespace: backyardsNamespace,
	}, "backyards")
	if err != nil {
		return nil, errors.WrapIf(err, "could not render helm manifest objects")
	}

	return objects, nil
}

func (c *installCommand) validate(options *installOptions) error {
	var istioHealthy bool
	var combinedErr error

	istioExists, istioHealthy, err := c.istioRunning(options.istioNamespace)
	if err != nil {
		return errors.WrapIf(err, "failed to check Istio state")
	}

	if !istioExists {
		combinedErr = errors.Combine(combinedErr,
			errors.Errorf("could not find Istio sidecar injector in '%s' namespace, " +
				"use the --install-istio flag", options.istioNamespace))
	}
	if istioExists && !istioHealthy {
		combinedErr = errors.Combine(combinedErr,
			errors.Errorf("Istio sidecar injector not healthy yet in '%s' namespace", options.istioNamespace))
	}

	if !options.disableCertManager {
		certManagerExists, certManagerHealthy, err := c.certManagerRunning()
		if err != nil {
			return errors.WrapIf(err, "failed to check cert-manager state")
		}

		if !certManagerExists {
			combinedErr = errors.Combine(combinedErr,
				errors.Errorf("could not find cert-manager controller in '%s' namespace, " +
					"use the --install-cert-manager flag or disable it using --disable-cert-manager " +
					"which disables dependent services as well", certmanager.CertManagerNamespace))
		}
		if certManagerExists && !certManagerHealthy {
			combinedErr = errors.Combine(combinedErr,
				errors.Errorf("cert-manager controller not healthy yet in '%s' namespace", certmanager.CertManagerNamespace))
		}
	}

	if options.disableCertManager && !options.disableAuditSink {
		combinedErr = errors.Combine(combinedErr, errors.Errorf("The HTTP AuditSink feature cannot work without cert-manager"))
	}

	return combinedErr
}

func (c *installCommand) istioRunning(istioNamespace string) (exists bool, healthy bool, err error) {
	cl, err := c.cli.GetK8sClient()
	if err != nil {
		err = errors.WrapIf(err, "could not get k8s client")
		return
	}
	var pods v1.PodList
	err = cl.List(context.Background(), &pods, client.InNamespace(istioNamespace), client.MatchingLabels(sidecarPodLabels))
	if err != nil {
		err = errors.WrapIf(err, "could not list istio pods")
		return
	}
	if len(pods.Items) > 0 {
		exists = true
	}
	for _, pod := range pods.Items {
		if pod.Status.Phase == v1.PodRunning {
			healthy = true
			break
		}
	}
	return
}

func (c *installCommand) certManagerRunning() (exists bool, healthy bool, err error) {
	cl, err := c.cli.GetK8sClient()
	if err != nil {
		err = errors.WrapIf(err, "could not get k8s client")
		return
	}
	var certManagerPods v1.PodList
	err = cl.List(context.Background(), &certManagerPods, client.InNamespace(certmanager.CertManagerNamespace),
		client.MatchingLabels(certManagerPodLabels))
	if err != nil {
		err = errors.WrapIf(err, "failed to list cert-manager controller pods")
		return
	}
	if len(certManagerPods.Items) > 0 {
		exists = true
	}
	for _, pod := range certManagerPods.Items {
		if pod.Status.Phase == v1.PodRunning {
			healthy = true
			break
		}
	}
	return
}

func (c *installCommand) runDemo(cli cli.CLI, options *installOptions) error {
	var err error

	if !options.runDemo || (!options.installEverything && !options.installDemoapp) {
		return nil
	}

	scmdOptions := demoapp.NewLoadOptions()
	scmdOptions.Nowait = true
	scmd := demoapp.NewLoadCommand(cli, scmdOptions)
	err = scmd.RunE(scmd, nil)
	if err != nil {
		return errors.WrapIf(err, "error during sending load to demo application")
	}

	dbOptions := NewDashboardOptions()
	dbOptions.URI = "?namespaces=" + demoapp.GetNamespace()
	dbOptions.Port = 0
	dbCmd := newDashboardCommand(cli, dbOptions)
	err = dbCmd.RunE(dbCmd, nil)
	if err != nil {
		return errors.WrapIf(err, "error during opening dashboard")
	}

	return nil
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

	if !options.disableCertManager && (options.installCertManager || options.installEverything) {
		scmdOptions := certmanager.NewInstallOptions()
		if options.dumpResources {
			scmdOptions.DumpResources = true
		}
		scmd = certmanager.NewInstallCommand(cli, scmdOptions)
		err = scmd.RunE(scmd, nil)
		if err != nil {
			return errors.WrapIf(err, "error during cert-manager install")
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
