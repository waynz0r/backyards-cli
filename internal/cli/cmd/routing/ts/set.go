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

package ts

import (
	"emperror.dev/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/routing/common"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/graphql"
)

type setCommand struct{}

type setOptions struct {
	serviceID string
	subsets   []string

	serviceName   types.NamespacedName
	parsedSubsets parsedSubsets
}

func newSetOptions() *setOptions {
	return &setOptions{}
}

func newSetCommand(cli cli.CLI) *cobra.Command {
	c := &setCommand{}
	options := newSetOptions()

	cmd := &cobra.Command{
		Use:           "set [[--service=]namespace/servicename] [[--version=]subset=weight] ...",
		Short:         "Set traffic shifting rules for a service",
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			if len(args) > 0 {
				options.serviceID = args[0]
			}

			if len(args) > 1 {
				options.subsets = append(options.subsets, args[1:]...)
			}

			if options.serviceID == "" {
				return errors.New("service must be specified")
			}

			if len(options.subsets) < 1 {
				return errors.New("at least 1 subset must be specified")
			}

			options.serviceName, err = common.ParseServiceID(options.serviceID)
			if err != nil {
				return err
			}

			options.parsedSubsets, err = parseSubsets(options.subsets)
			if err != nil {
				return err
			}

			return c.run(cli, options)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&options.serviceID, "service", "", "Service name")
	flags.StringArrayVar(&options.subsets, "subset", []string{}, "Subsets with weights (sum of the weight must add up to 100)")

	return cmd
}

func (c *setCommand) run(cli cli.CLI, options *setOptions) error {
	var err error

	service, err := common.GetServiceByName(cli, options.serviceName)
	if err != nil {
		if k8serrors.IsNotFound(errors.Cause(err)) {
			return err
		}
		return errors.WrapIf(err, "could not get service")
	}

	client, err := common.GetGraphQLClient(cli)
	if err != nil {
		return errors.WrapIf(err, "could not get initialized graphql client")
	}

	req := graphql.ApplyHTTPRouteRequest{
		Name:      service.Name,
		Namespace: service.Namespace,
		Route:     make([]graphql.HTTPRouteDestination, 0),
	}

	for subset, weight := range options.parsedSubsets {
		req.Route = append(req.Route, graphql.HTTPRouteDestination{
			Destination: graphql.Destination{
				Host:   service.Name,
				Subset: subset,
			},
			Weight: weight,
		})
	}

	r, err := client.ApplyHTTPRoute(req)
	if err != nil {
		return err
	}

	if !r {
		return errors.New("unknown error: cannot set traffic shifting")
	}

	log.Infof("traffic shifting for %s set to %s successfully", options.serviceName, options.parsedSubsets)

	return nil
}
