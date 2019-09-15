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
	"os"
	"regexp"

	"emperror.dev/errors"
	logrushandler "emperror.dev/handler/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/cli/cmd/canary"
	"github.com/banzaicloud/backyards-cli/pkg/cli/cmd/certmanager"
	"github.com/banzaicloud/backyards-cli/pkg/cli/cmd/demoapp"
	"github.com/banzaicloud/backyards-cli/pkg/cli/cmd/istio"
	"github.com/banzaicloud/backyards-cli/pkg/cli/cmd/routing"
)

const (
	defaultNamespace = "backyards-system"
)

var (
	backyardsNamespace string
	kubeconfigPath     string
	kubeContext        string
	verbose            bool
	outputFormat       string

	namespaceRegex = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
)

// RootCmd represents the root Cobra command
var RootCmd = &cobra.Command{
	Use:           "backyards",
	Short:         "Install and manage Backyards",
	SilenceErrors: true,
	SilenceUsage:  false,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		if verbose {
			log.SetLevel(log.DebugLevel)
		} else {
			log.SetLevel(log.InfoLevel)
		}

		namespaceFromEnv := os.Getenv("BACKYARDS_NAMESPACE")
		if backyardsNamespace == defaultNamespace && namespaceFromEnv != "" {
			backyardsNamespace = namespaceFromEnv
		}

		if !namespaceRegex.MatchString(backyardsNamespace) {
			return errors.NewWithDetails("invalid namespace", "namespace", backyardsNamespace)
		}

		return nil
	},
}

// Init is a temporary function to set initial values in the root cmd
func Init(version string, commitHash string, buildDate string) {
	RootCmd.Version = version

	RootCmd.SetVersionTemplate(fmt.Sprintf(
		"Backyards CLI version %s (%s) built on %s\n",
		version,
		commitHash,
		buildDate,
	))
}

// GetRootCommand returns the cli root command
func GetRootCommand() *cobra.Command {
	return RootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately
// This is called by main.main(). It only needs to happen once to the RootCmd
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		handler := logrushandler.New(log.New())
		handler.Handle(err)
		os.Exit(1)
	}
}

func init() {
	flags := RootCmd.PersistentFlags()
	flags.StringVarP(&backyardsNamespace, "namespace", "n", defaultNamespace, "namespace in which Backyards is installed [$BACKYARDS_NAMESPACE]")
	_ = viper.BindPFlag("backyards.namespace", flags.Lookup("namespace"))
	flags.StringVarP(&kubeconfigPath, "kubeconfig", "c", "", "path to the kubeconfig file to use for CLI requests")
	_ = viper.BindPFlag("kubeconfig", flags.Lookup("kubeconfig"))
	flags.StringVar(&kubeContext, "context", "", "name of the kubeconfig context to use")
	_ = viper.BindPFlag("kubecontext", flags.Lookup("context"))
	flags.BoolVarP(&verbose, "verbose", "v", false, "turn on debug logging")

	flags.StringVarP(&outputFormat, "output", "o", "table", "output format (table|yaml|json)")
	_ = viper.BindPFlag("output.format", flags.Lookup("output"))

	_ = viper.BindPFlag("formatting.force-color", flags.Lookup("color"))
	flags.Bool("non-interactive", false, "never ask questions interactively")
	_ = viper.BindPFlag("formatting.non-interactive", flags.Lookup("non-interactive"))
	flags.Bool("interactive", false, "ask questions interactively even if stdin or stdout is non-tty")
	_ = viper.BindPFlag("formatting.force-interactive", flags.Lookup("interactive"))

	cli := cli.NewCli(os.Stdout)

	RootCmd.AddCommand(newVersionCommand(cli))
	RootCmd.AddCommand(newInstallCommand(cli))
	RootCmd.AddCommand(newUninstallCommand(cli))
	RootCmd.AddCommand(newDashboardCommand(cli, NewDashboardOptions()))
	RootCmd.AddCommand(istio.NewRootCmd(cli))
	RootCmd.AddCommand(canary.NewRootCmd(cli))
	RootCmd.AddCommand(demoapp.NewRootCmd(cli))
	RootCmd.AddCommand(routing.NewRootCmd(cli))
	RootCmd.AddCommand(certmanager.NewRootCmd(cli))
	RootCmd.AddCommand(NewGraphCmd(cli, "graph", "base.json"))
	RootCmd.AddCommand(NewGraphCmd(cli, "cb", "cb.json"))
}
