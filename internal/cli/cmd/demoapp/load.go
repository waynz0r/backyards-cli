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
	"sync"

	"emperror.dev/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/graphql"
)

type loadCommand struct {
	cli cli.CLI
}

type LoadOptions struct {
	Nowait    bool
	Frequency int
	Duration  int

	namespace string
}

func NewLoadOptions() *LoadOptions {
	return &LoadOptions{
		Nowait: false,

		Frequency: 10,
		Duration:  30,
	}
}

func NewLoadCommand(cli cli.CLI, options *LoadOptions) *cobra.Command {
	c := &loadCommand{
		cli: cli,
	}

	cmd := &cobra.Command{
		Use:   "load",
		Args:  cobra.NoArgs,
		Short: "Send load to demo application",

		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			if options.namespace == "" {
				options.namespace = backyardsDemoNamespace
			}

			return c.run(cli, options)
		},
	}

	cmd.Flags().IntVar(&options.Frequency, "rps", options.Frequency, "Number of requests per second")
	cmd.Flags().IntVar(&options.Duration, "duration", options.Duration, "Duration in seconds")

	return cmd
}

func (c *loadCommand) run(cli cli.CLI, options *LoadOptions) error {
	var err error
	var response graphql.GenerateLoadResponse

	pf, err := cli.GetPortforwardForIGW(0)
	if err != nil {
		return err
	}

	err = pf.Run()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	log.WithFields(log.Fields{
		"rps":      options.Frequency,
		"duration": options.Duration,
	}).Info("sending load to demo application")
	go func() {
		client := graphql.NewClient(pf.GetURL("/api/graphql"))
		response, err = client.GenerateLoad(graphql.GenerateLoadRequest{
			Namespace: options.namespace,
			Service:   "frontpage",
			Port:      8080,
			Endpoint:  "/",
			Method:    "GET",
			Frequency: options.Frequency,
			Duration:  options.Duration,
			Headers:   nil,
		})

		log.Info("loader stopped")
		for code, count := range response {
			log.WithFields(log.Fields{
				"responseCode": code,
				"requestCount": count,
			}).Info("")
		}

		wg.Done()
	}()

	if !options.Nowait {
		wg.Wait()
	}

	if err != nil {
		return errors.WrapIf(err, "error during load generation")
	}

	return nil
}
