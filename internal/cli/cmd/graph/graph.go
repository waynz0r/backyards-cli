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
package graph

import (
	"context"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"

	"github.com/waynz0r/grafterm/pkg/controller"
	"github.com/waynz0r/grafterm/pkg/model"
	"github.com/waynz0r/grafterm/pkg/service/configuration"
	"github.com/waynz0r/grafterm/pkg/service/log"
	"github.com/waynz0r/grafterm/pkg/service/metric"
	metricdatasource "github.com/waynz0r/grafterm/pkg/service/metric/datasource"
	metricmiddleware "github.com/waynz0r/grafterm/pkg/service/metric/middleware"
	"github.com/waynz0r/grafterm/pkg/view"
	"github.com/waynz0r/grafterm/pkg/view/page"
	"github.com/waynz0r/grafterm/pkg/view/render"
	"github.com/waynz0r/grafterm/pkg/view/render/termdash"

	"github.com/banzaicloud/backyards-cli/cmd/backyards/static/graphtemplates"
	"github.com/banzaicloud/backyards-cli/internal/cli/cmd/routing/common"
	"github.com/banzaicloud/backyards-cli/pkg/cli"
)

var (
	titleSuffix     string
	outbound        bool
	refreshInterval time.Duration
	relativeDur     time.Duration
)

type graphOptions struct {
	serviceID string

	serviceName types.NamespacedName
}

func newGraphOptions() *graphOptions {
	return &graphOptions{}
}

func NewGraphCmd(cli cli.CLI, fileName string) *cobra.Command {
	options := newGraphOptions()

	cmd := &cobra.Command{
		Use:   "graph [[--service=]namespace/servicename]",
		Short: "Show graph",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			var err error

			if len(args) > 0 {
				options.serviceID = args[0]
			}

			if options.serviceID != "" {
				options.serviceName, err = common.ParseServiceID(options.serviceID)
				if err != nil {
					return err
				}
			}

			f, err := graphtemplates.GraphTemplates.Open(fileName)
			if err != nil {
				return err
			}

			cfg, err := configuration.JSONLoader{}.Load(f)
			if err != nil {
				return err
			}

			ddss, err := cfg.Datasources()
			if err != nil {
				return err
			}

			pf, err := cli.GetPortforwardForIGW(0)
			if err != nil {
				return err
			}

			err = pf.Run()
			if err != nil {
				return err
			}

			udss := []model.Datasource{
				{
					ID: "ds",
					DatasourceSource: model.DatasourceSource{
						Prometheus: &model.PrometheusDatasource{
							Address: pf.GetURL("/prometheus"),
						},
					},
				},
			}

			gatherer, err := createGatherer(ddss, udss)
			if err != nil {
				return err
			}

			// Create controller.
			ctrl := controller.NewController(gatherer)

			// Create renderer.
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			renderer, err := termdash.NewTermDashboard(cancel, log.Dummy)
			if err != nil {
				return err
			}
			defer renderer.Close()

			appcfg := view.AppConfig{
				RefreshInterval:   refreshInterval,
				RelativeTimeRange: relativeDur,
			}

			ds, err := cfg.Dashboard()
			if err != nil {
				return err
			}

			app, err := createApp(ctx, appcfg, ds, ctrl, renderer, options)
			if err != nil {
				return err
			}

			err = app.Run(ctx)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&titleSuffix, "title-suffix", "", "Title suffix")
	cmd.Flags().BoolVar(&outbound, "outbound", false, "Whether to show outbound or inbound metrics")
	cmd.Flags().DurationVarP(&refreshInterval, "refresh-interval", "r", 10*time.Second, "the interval to refresh the dashboard")
	cmd.Flags().DurationVarP(&relativeDur, "relative-duration", "d", 15*time.Minute, "the relative duration from now to load the graph")

	return cmd
}

func getFilter(options *graphOptions) string {
	filters := make([]string, 0)

	if outbound {
		filters = append(filters, "reporter=\"source\"")
	} else {
		filters = append(filters, "reporter=\"destination\"")
	}

	if options.serviceName.Namespace != "" {
		filters = append(filters, "destination_service_namespace=~\""+options.serviceName.Namespace+"\"")
	}

	if options.serviceName.Name != "" {
		filters = append(filters, "destination_service_name=~\""+options.serviceName.Name+"\"")
	}

	return strings.Join(filters, ",")
}

func getTitleSuffix(options *graphOptions) string {
	if titleSuffix != "" {
		return titleSuffix
	}

	s := make([]string, 0)
	if outbound {
		s = append(s, "outbound")
	} else {
		s = append(s, "inbound")
	}

	if options.serviceName.Namespace != "" {
		s = append(s, options.serviceName.Namespace)
	}

	if options.serviceName.Name != "" {
		s = append(s, options.serviceName.Name)
	}

	return strings.Join(s, " / ")
}

func createGatherer(dashboardDss, userDss []model.Datasource) (metric.Gatherer, error) {
	gatherer, err := metricdatasource.NewGatherer(metricdatasource.ConfigGatherer{
		DashboardDatasources: dashboardDss,
		UserDatasources:      userDss,
	})
	if err != nil {
		return nil, err
	}
	gatherer = metricmiddleware.Logger(log.Dummy, gatherer)

	return gatherer, nil
}

func createApp(ctx context.Context, appCfg view.AppConfig, dashboard model.Dashboard, ctrl controller.Controller, renderer render.Renderer, options *graphOptions) (*view.App, error) {

	filter := getFilter(options)
	titleSuffix = " " + getTitleSuffix(options)

	dashCfg := page.DashboardCfg{
		AppRelativeTimeRange: relativeDur,
		AppOverrideVariables: map[string]string{
			"titleSuffix": titleSuffix,
			"filter":      filter,
		},
		Controller: ctrl,
		Dashboard:  dashboard,
		Renderer:   renderer,
	}

	syncer, err := page.NewDashboard(ctx, dashCfg, log.Dummy)
	if err != nil {
		return nil, err
	}
	app := view.NewApp(appCfg, syncer, log.Dummy)
	return app, nil
}
