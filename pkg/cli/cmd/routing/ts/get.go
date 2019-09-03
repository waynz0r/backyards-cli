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
)

type getCommand struct{}

type getOptions struct {
	serviceID string

	serviceName types.NamespacedName
}

func newGetOptions() *getOptions {
	return &getOptions{}
}

func newGetCommand(cli cli.CLI) *cobra.Command {
	c := &getCommand{}
	options := newGetOptions()

	cmd := &cobra.Command{
		Use:           "get [[--service=]namespace/servicename]",
		Short:         "Get traffic shifting rules for a service",
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

			options.serviceName, err = parseServiceID(options.serviceID)
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

func (c *getCommand) run(cli cli.CLI, options *getOptions) error {
	var err error

	_, err = getService(cli, options.serviceName)
	if err != nil {
		if k8serrors.IsNotFound(errors.Cause(err)) {
			return err
		}
		return errors.WrapIf(err, "could not get service")
	}

	vservice, err := getVirtualserviceByName(cli, options.serviceName)
	if err != nil {
		if k8serrors.IsNotFound(errors.Cause(err)) {
			log.Infof("no traffic shifting rules set for %s", options.serviceName)
			return nil
		}
		return errors.WrapIf(err, "could not get service")
	}

	subsets := make(parsedSubsets, 0)
	for _, route := range vservice.Spec.HTTP {
		if len(route.Match) > 0 {
			continue
		}

		for _, r := range route.Route {
			subsets[r.Destination.Subset] = r.Weight
		}
	}

	log.Infof("traffic shifting for %s is currently set to %s", options.serviceName, subsets)

	return nil
}
