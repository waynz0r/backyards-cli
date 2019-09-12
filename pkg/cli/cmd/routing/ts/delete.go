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

	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/cli/cmd/routing/common"
	"github.com/banzaicloud/backyards-cli/pkg/graphql"
)

type deleteCommand struct{}

type deleteOptions struct {
	serviceID string

	serviceName types.NamespacedName
}

func newDeleteOptions() *deleteOptions {
	return &deleteOptions{}
}

func newDeleteCommand(cli cli.CLI) *cobra.Command {
	c := &deleteCommand{}
	options := newDeleteOptions()

	cmd := &cobra.Command{
		Use:           "delete [[--service=]namespace/servicename]",
		Short:         "Delete traffic shifting rules of a service",
		Args:          cobra.MaximumNArgs(1),
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			if len(args) > 0 {
				options.serviceID = args[0]
			}

			if options.serviceID == "" {
				return errors.New("service must be specified")
			}

			options.serviceName, err = common.ParseServiceID(options.serviceID)
			if err != nil {
				return err
			}

			return c.run(cli, options)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&options.serviceID, "service", "", "Service name")

	return cmd
}

func (c *deleteCommand) run(cli cli.CLI, options *deleteOptions) error {
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

	req := graphql.DisableHTTPRouteRequest{
		Name:      service.Name,
		Namespace: service.Namespace,
		Rules: []string{
			"Route",
		},
	}
	r, err := client.DisableHTTPRoute(req)
	if err != nil {
		return err
	}

	if !r {
		return errors.New("unknown error: cannot delete traffic shifting")
	}

	log.Infof("traffic shifting rules set to %s successfully deleted", options.serviceName)

	return nil
}
