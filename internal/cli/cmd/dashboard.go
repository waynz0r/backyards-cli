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
	defaultLocalPort = 50500
)

type dashboardCommand struct{}

type DashboardOptions struct {
	URI  string
	Port int
	wait time.Duration
}

func NewDashboardOptions() *DashboardOptions {
	return &DashboardOptions{
		URI:  "",
		Port: defaultLocalPort,
		wait: 300 * time.Second,
	}
}

func NewDashboardCommand(cli cli.CLI, options *DashboardOptions) *cobra.Command {
	c := dashboardCommand{}

	cmd := &cobra.Command{
		Use:   "dashboard [flags]",
		Short: "Open the Backyards dashboard in a web browser",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if options.Port < 0 {
				log.Error(errors.NewWithDetails("port must be greater than or equal to zero", "port", options.Port))
				return nil
			}

			err := c.run(cli, options)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.PersistentFlags().IntVarP(&options.Port, "port", "p", options.Port, "The local port on which to serve requests (when set to 0, a random port will be used)")

	return cmd
}

func (c *dashboardCommand) run(cli cli.CLI, options *DashboardOptions) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	defer signal.Stop(signals)

	pf, err := cli.GetPortforwardForIGW(options.Port)
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

	url := pf.GetURL(options.URI)
	log.Infof("Backyards UI is available at %s", url)
	err = browser.OpenURL(url)
	if err != nil {
		return err
	}

	pf.WaitForStop()

	return nil
}
