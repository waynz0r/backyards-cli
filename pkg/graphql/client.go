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

package graphql

import (
	"github.com/machinebox/graphql"
)

type Client interface {
	SetJWTToken(string)
	GenerateLoad(req GenerateLoadRequest) (GenerateLoadResponse, error)
	ApplyHTTPRoute(req ApplyHTTPRouteRequest) (ApplyHTTPRouteResponse, error)
	DisableHTTPRoute(req DisableHTTPRouteRequest) (DisableHTTPRouteResponse, error)
	ApplyGlobalTrafficPolicy(req ApplyGlobalTrafficPolicyRequest) (ApplyGlobalTrafficPolicyResponse, error)
	DisableGlobalTrafficPolicy(req DisableGlobalTrafficPolicyRequest) (DisableGlobalTrafficPolicyResponse, error)
}

type client struct {
	jwtToken string
	client   *graphql.Client
}

func NewClient(url string, opt ...graphql.ClientOption) Client {
	return &client{
		client: graphql.NewClient(url, opt...),
	}
}

func (c *client) SetJWTToken(token string) {
	c.jwtToken = token
}
