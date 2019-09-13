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
	"context"

	"github.com/MakeNowJust/heredoc"
)

type HTTPRouteDestination struct {
	Destination Destination `json:"destination"`
	Weight      int         `json:"weight"`
}

type Destination struct {
	Host   string        `json:"host"`
	Subset string        `json:"subset,omitempty"`
	Port   *PortSelector `json:"port,omitempty"`
}

type PortSelector struct {
	Number uint32 `json:"number,omitempty"`
	Name   string `json:"name,omitempty"`
}

type ApplyHTTPRouteRequest struct {
	Name      string                 `json:"name"`
	Namespace string                 `json:"namespace"`
	Route     []HTTPRouteDestination `json:"route,omitempty"`
}

type ApplyHTTPRouteResponse bool

func (c *client) ApplyHTTPRoute(req ApplyHTTPRouteRequest) (ApplyHTTPRouteResponse, error) {
	request := heredoc.Doc(`
	  mutation applyHTTPRoute(
		$input: ApplyHTTPRouteInput!
	  ) {
		applyHTTPRoute(
		  input: $input
		)
	  }
`)

	r := c.NewRequest(request)
	r.Var("input", req)

	// run it and capture the response
	var respData map[string]ApplyHTTPRouteResponse
	if err := c.client.Run(context.Background(), r, &respData); err != nil {
		return false, err
	}

	return respData["applyHTTPRoute"], nil
}
