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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/types"

	"github.com/banzaicloud/backyards-cli/pkg/cli"
	"github.com/banzaicloud/backyards-cli/pkg/graphql"
	"github.com/banzaicloud/backyards-cli/pkg/k8s"
)

const (
	backyardsServiceAccountName = "backyards"
)

func NewRootCmd(cli cli.CLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "traffic-shifting",
		Aliases: []string{"ts"},
		Short:   "Manage traffic-shifting configurations",
	}

	cmd.AddCommand(
		newGetCommand(cli),
		newSetCommand(cli),
		newDeleteCommand(cli),
	)

	return cmd
}

func getGraphQLClient(cli cli.CLI) (graphql.Client, error) {
	k8sclient, err := cli.GetK8sClient()
	if err != nil {
		return nil, err
	}

	token, err := k8s.GetTokenForServiceAccountName(k8sclient, types.NamespacedName{
		Name:      backyardsServiceAccountName,
		Namespace: viper.GetString("backyards.namespace"),
	})
	if err != nil {
		return nil, err
	}

	pf, err := cli.GetPortforwardForIGW(0)
	if err != nil {
		return nil, err
	}

	err = pf.Run()
	if err != nil {
		return nil, err
	}

	client := graphql.NewClient(pf.GetURL("/api/graphql"))
	client.SetJWTToken(token)

	return client, nil
}
