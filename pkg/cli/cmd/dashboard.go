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
	"os"
	"os/signal"
	"time"

	"emperror.dev/errors"
	"github.com/pkg/browser"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/banzaicloud/backyards-cli/pkg/cli"
)

var (
	IGWPort        = 80
	IGWMatchLabels = map[string]string{
		"app.kubernetes.io/component": "ingressgateway",
		"app.kubernetes.io/instance":  "backyards",
	}
	defaultLocalPort = 50500
)

type dashboardCommand struct{}

type dashboardOptions struct {
	port int
	wait time.Duration
}

func newDashboardOptions() *dashboardOptions {
	return &dashboardOptions{
		port: defaultLocalPort,
		wait: 300 * time.Second,
	}
}

func newDashboardCommand(cli cli.CLI) *cobra.Command {
	c := dashboardCommand{}
	options := newDashboardOptions()

	cmd := &cobra.Command{
		Use:   "dashboard [flags]",
		Short: "Open the Backyards dashboard in a web browser",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if options.port < 0 {
				log.Error(errors.NewWithDetails("port must be greater than or equal to zero", "port", options.port))
				return nil
			}

			err := c.run(cli, options)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.PersistentFlags().IntVarP(&options.port, "port", "p", options.port, "The local port on which to serve requests (when set to 0, a random port will be used)")

	return cmd
}

func (c *dashboardCommand) run(cli cli.CLI, options *dashboardOptions) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	defer signal.Stop(signals)

	pf, err := cli.GetPortforwardForPod(IGWMatchLabels, backyardsNamespace, options.port, IGWPort)
	if err != nil {
		return errors.WrapIf(err, "cloud not create port forwarder")
	}

	go func() {
		<-signals
		pf.Stop()
	}()

	err = pf.Run()
	if err != nil {
		return errors.WrapIf(err, "could not run port forwarder")
	}

	url := pf.GetURL("")
	log.Infof("Backyards UI is available at %s", url)
	browser.OpenURL(url)

	pf.WaitForStop()

	return nil
}
