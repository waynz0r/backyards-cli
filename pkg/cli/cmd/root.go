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
	"github.com/banzaicloud/backyards-cli/pkg/cli/cmd/demoapp"
	"github.com/banzaicloud/backyards-cli/pkg/cli/cmd/istio"
)

const (
	defaultNamespace = "backyards-system"
)

var (
	backyardsNamespace string
	kubeconfigPath     string
	kubeContext        string
	verbose            bool

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
	RootCmd.PersistentFlags().StringVarP(&backyardsNamespace, "namespace", "n", defaultNamespace, "Namespace in which Backyards is installed [$BACKYARDS_NAMESPACE]")
	viper.BindPFlag("backyards.namespace", RootCmd.PersistentFlags().Lookup("namespace"))
	RootCmd.PersistentFlags().StringVarP(&kubeconfigPath, "kubeconfig", "c", "", "Path to the kubeconfig file to use for CLI requests")
	viper.BindPFlag("kubeconfig", RootCmd.PersistentFlags().Lookup("kubeconfig"))
	RootCmd.PersistentFlags().StringVar(&kubeContext, "context", "", "Name of the kubeconfig context to use")
	viper.BindPFlag("kubecontext", RootCmd.PersistentFlags().Lookup("context"))
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Turn on debug logging")

	cli := cli.NewCli(os.Stdout)

	RootCmd.AddCommand(newVersionCommand(cli))
	RootCmd.AddCommand(newInstallCommand(cli))
	RootCmd.AddCommand(newUninstallCommand(cli))
	RootCmd.AddCommand(newDashboardCommand(cli, NewDashboardOptions()))
	RootCmd.AddCommand(istio.NewRootCmd(cli))
	RootCmd.AddCommand(canary.NewRootCmd(cli))
	RootCmd.AddCommand(demoapp.NewRootCmd(cli))
}
