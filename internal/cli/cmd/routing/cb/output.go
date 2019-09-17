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

	"github.com/banzaicloud/backyards-cli/pkg/output"
)

func Output(cli output.FormatContext, data interface{}) error {
	ctx := &output.Context{
		Out:     cli.Out(),
		Color:   cli.Color(),
		Format:  cli.OutputFormat(),
		Fields:  []string{"MaxConnections", "ConnectTimeout", "HTTP1MaxPendingRequests", "HTTP2MaxRequests", "MaxRequestsPerConnection", "MaxRetries", "ConsecutiveErrors", "Interval", "BaseEjectionTime", "MaxEjectionPercent"},
		Headers: []string{"Connections", "Timeout", "Pending Requests", "Requests", "RPC", "Retries", "Errors", "Interval", "Ejection time", "percentage"},
	}

	err := output.Output(ctx, data)
	if err != nil {
		return errors.WrapIf(err, "could not produce output")
	}

	return nil
}
