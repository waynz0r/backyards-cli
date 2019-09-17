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
	"time"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis/istio/v1alpha3"

	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/routing/common"
	clierrors "github.com/banzaicloud/backyards-cli/internal/errors"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/graphql"
	"github.com/banzaicloud/backyards-cli/pkg/questionnaire"
)

type setCommand struct{}

type CircuitBreakerSettings struct {
	// TCP
	MaxConnections int32  `json:"maxConnections,omitempty" yaml:"maxConnections" survey.question:"Maximum number of HTTP1/TCP connections" survey.validate:"int"`
	ConnectTimeout string `json:"connectTimeout,omitempty" yaml:"connectTimeout" survey.question:"TCP connection timeout" survey.validate:"durationstring"`

	// HTTP
	HTTP1MaxPendingRequests  int32 `json:"http1MaxPendingRequests,omitempty" yaml:"http1MaxPendingRequests,omitempty" survey.question:"Maximum number of pending HTTP requests"`
	HTTP2MaxRequests         int32 `json:"http2MaxRequests,omitempty" yaml:"http2MaxRequests,omitempty" survey.question:"Maximum number of requests"`
	MaxRequestsPerConnection int32 `json:"maxRequestsPerConnection,omitempty" yaml:"maxRequestsPerConnection,omitempty" survey.question:"Maximum number of requests per connection"`
	MaxRetries               int32 `json:"maxRetries,omitempty" yaml:"maxRetries,omitempty" survey.question:"Maximum number of retries"`

	// Outlier
	ConsecutiveErrors  int32  `json:"consecutiveErrors,omitempty" yaml:"consecutiveErrors,omitempty" survey.question:"Number of errors before a host is ejected"`
	Interval           string `json:"interval,omitempty" yaml:"interval,omitempty" survey.question:"Time interval between ejection sweep analysis"`
	BaseEjectionTime   string `json:"baseEjectionTime,omitempty" yaml:"baseEjectionTime,omitempty" survey.question:"Minimum ejection duration"`
	MaxEjectionPercent int32  `json:"maxEjectionPercent,omitempty" yaml:"maxEjectionPercent,omitempty" survey.question:"Maximum ejection percentage"`
}

type setOptions struct {
	serviceID string

	CircuitBreakerSettings
	connectTimeout   time.Duration
	interval         time.Duration
	baseEjectionTime time.Duration

	serviceName types.NamespacedName
}

func newSetOptions() *setOptions {
	return &setOptions{
		CircuitBreakerSettings: CircuitBreakerSettings{
			MaxConnections:           1024,
			ConnectTimeout:           "3s",
			HTTP1MaxPendingRequests:  1024,
			HTTP2MaxRequests:         1024,
			MaxRequestsPerConnection: 1,
			MaxRetries:               1024,
			Interval:                 "10s",
			BaseEjectionTime:         "10s",
			ConsecutiveErrors:        5,
			MaxEjectionPercent:       100,
		},
		connectTimeout:   3 * time.Second,
		interval:         10 * time.Second,
		baseEjectionTime: 30 * time.Second,
	}
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

			if options.serviceID == "" {
				return errors.New("service must be specified")
			}

			options.ConnectTimeout = options.connectTimeout.String()
			options.BaseEjectionTime = options.baseEjectionTime.String()
			options.Interval = options.interval.String()

			options.serviceName, err = common.ParseServiceID(options.serviceID)
			if err != nil {
				return err
			}

			err = c.askQuestions(cli, options)
			if err != nil {
				return err
			}

			return c.run(cli, options)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&options.serviceID, "service", "", "Service name")

	// TCP
	flags.Int32Var(&options.MaxConnections, "max-connections", options.MaxConnections, "Maximum number of HTTP1/TCP connections to a destination host")
	flags.DurationVar(&options.connectTimeout, "connect-timeout", options.connectTimeout, "TCP connection timeout")

	// HTTP
	flags.Int32Var(&options.HTTP1MaxPendingRequests, "max-pending-requests", options.HTTP1MaxPendingRequests, "Maximum number of pending HTTP requests to a destination")
	flags.Int32Var(&options.HTTP2MaxRequests, "max-requests", options.HTTP2MaxRequests, "Maximum number of requests to a backend")
	flags.Int32Var(&options.MaxRequestsPerConnection, "max-requests-per-connection", options.MaxRequestsPerConnection, "Maximum number of requests per connection to a backend. Setting this parameter to 1 disables keep alive")
	flags.Int32Var(&options.MaxRetries, "max-retries", options.MaxRetries, "Maximum number of retries that can be outstanding to all hosts in an envoy cluster at a given time")

	// Outlier
	flags.Int32Var(&options.ConsecutiveErrors, "consecutiveErrors", options.ConsecutiveErrors, "Number of errors before a host is ejected from the connection pool")
	flags.DurationVar(&options.interval, "interval", options.interval, "Time interval between ejection sweep analysis")
	flags.DurationVar(&options.baseEjectionTime, "baseEjectionTime", options.baseEjectionTime, "Minimum ejection duration. A host will remain ejected for a period equal to the product of minimum ejection duration and the number of times the host has been ejected")
	flags.Int32Var(&options.MaxEjectionPercent, "maxEjectionPercent", options.MaxEjectionPercent, "Maximum % of hosts in the load balancing pool for the upstream service that can be ejected")

	return cmd
}

func (c *setCommand) askQuestions(cli cli.CLI, options *setOptions) error {
	var err error

	if !cli.InteractiveTerminal() {
		return nil
	}

	qs, err := questionnaire.GetQuestionsFromStruct(options.CircuitBreakerSettings)
	if err != nil {
		return err
	}

	err = survey.Ask(qs, &options.CircuitBreakerSettings)
	if err != nil {
		return errors.Wrap(err, "error while asking question")
	}

	return nil
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

	req := graphql.ApplyGlobalTrafficPolicyRequest{
		Name:      service.Name,
		Namespace: service.Namespace,
		ConnectionPool: &v1alpha3.ConnectionPoolSettings{
			TCP: &v1alpha3.TCPSettings{
				MaxConnections: options.MaxConnections,
				ConnectTimeout: options.ConnectTimeout,
			},
			HTTP: &v1alpha3.HTTPSettings{
				HTTP1MaxPendingRequests:  options.HTTP1MaxPendingRequests,
				HTTP2MaxRequests:         options.HTTP2MaxRequests,
				MaxRequestsPerConnection: options.MaxRequestsPerConnection,
				MaxRetries:               options.MaxRetries,
			},
		},
		OutlierDetection: &v1alpha3.OutlierDetection{
			ConsecutiveErrors:  options.ConsecutiveErrors,
			Interval:           options.Interval,
			BaseEjectionTime:   options.BaseEjectionTime,
			MaxEjectionPercent: options.MaxEjectionPercent,
		},
	}

	r, err := client.ApplyGlobalTrafficPolicy(req)
	if err != nil {
		return err
	}

	if !r {
		return errors.New("unknown error: cannot apply circuit breaker settings")
	}

	err = c.output(cli, options)
	if err != nil {
		return err
	}

	return nil
}

func (c *setCommand) output(cli cli.CLI, options *setOptions) error {
	data, err := getCircuitBreakerRulesByServiceName(cli, options.serviceName)
	if err != nil {
		if clierrors.IsNotFound(err) {
			log.Infof("no circuit breaker rules set for '%s'", options.serviceName)
			return nil
		}
		return err
	}

	if cli.InteractiveTerminal() {
		log.Infof("circuit breaker rules successfully applied to '%s'", options.serviceName)
	}

	err = Output(cli, data)
	if err != nil {
		return err
	}

	return nil
}
