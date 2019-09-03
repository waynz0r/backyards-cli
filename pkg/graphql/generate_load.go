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

	"github.com/machinebox/graphql"
)

type GenerateLoadRequest struct {
	Namespace string
	Service   string
	Port      int
	Endpoint  string
	Method    string
	Frequency int
	Duration  int
	Headers   map[string]string
}

type GenerateLoadResponse map[string]int

func (c *client) GenerateLoad(req GenerateLoadRequest) (GenerateLoadResponse, error) {
	request := `
	mutation load($namespace: String!, $service: String!, $port: Int!, $endpoint: String!, $method: String!, $body: String, $headers: Map, $frequency: Int!, $duration: Int!) {
		generateLoad(namespace: $namespace, service: $service, port: $port, endpoint: $endpoint, method: $method, body: $body, headers: $headers, frequency: $frequency, duration: $duration)
	}
`

	r := graphql.NewRequest(request)

	r.Var("namespace", req.Namespace)
	r.Var("service", req.Service)
	r.Var("port", req.Port)
	r.Var("endpoint", req.Endpoint)
	r.Var("method", req.Method)
	r.Var("frequency", req.Frequency)
	r.Var("duration", req.Duration)
	r.Var("headers", req.Headers)

	// set header fields
	r.Header.Set("Cache-Control", "no-cache")

	// run it and capture the response
	var respData map[string]GenerateLoadResponse
	if err := c.client.Run(context.Background(), r, &respData); err != nil {
		return nil, err
	}

	resp := respData["generateLoad"]

	return resp, nil
}
