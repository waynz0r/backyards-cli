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

package cb

import (
	"emperror.dev/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	clierrors "github.com/banzaicloud/backyards-cli/internal/errors"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/cli/cmd/routing/common"
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
		Short:         "Get circuit breaker rules for a service",
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

func getCircuitBreakerRulesByServiceName(cli cli.CLI, serviceName types.NamespacedName) (*CircuitBreakerSettings, error) {
	var err error

	_, err = common.GetServiceByName(cli, serviceName)
	if err != nil {
		if k8serrors.IsNotFound(errors.Cause(err)) {
			return nil, err
		}
		return nil, errors.WrapIf(err, "could not get service")
	}

	drule, err := common.GetDestinationRuleByName(cli, serviceName)
	notfound := false
	if err != nil {
		if k8serrors.IsNotFound(errors.Cause(err)) {
			notfound = true
		} else {
			return nil, errors.WrapIf(err, "could not get service")
		}
	} else if drule.Spec.TrafficPolicy == nil || drule.Spec.TrafficPolicy.ConnectionPool == nil {
		notfound = true
	}

	if notfound {
		return nil, clierrors.NotFoundError{}
	}

	tp := drule.Spec.TrafficPolicy

	return &CircuitBreakerSettings{
		MaxConnections: tp.ConnectionPool.TCP.MaxConnections,
		ConnectTimeout: tp.ConnectionPool.TCP.ConnectTimeout,

		HTTP1MaxPendingRequests:  tp.ConnectionPool.HTTP.HTTP1MaxPendingRequests,
		HTTP2MaxRequests:         tp.ConnectionPool.HTTP.HTTP2MaxRequests,
		MaxRequestsPerConnection: tp.ConnectionPool.HTTP.MaxRequestsPerConnection,
		MaxRetries:               tp.ConnectionPool.HTTP.MaxRetries,

		ConsecutiveErrors:  tp.OutlierDetection.ConsecutiveErrors,
		Interval:           tp.OutlierDetection.Interval,
		BaseEjectionTime:   tp.OutlierDetection.BaseEjectionTime,
		MaxEjectionPercent: tp.OutlierDetection.MaxEjectionPercent,
	}, nil
}

func (c *getCommand) run(cli cli.CLI, options *getOptions) error {
	var err error

	data, err := getCircuitBreakerRulesByServiceName(cli, options.serviceName)
	if err != nil {
		if clierrors.IsNotFound(err) {
			log.Infof("no circuit breaker rules set for %s", options.serviceName)
			return nil
		}
		return err
	}

	err = Output(cli, data)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}
