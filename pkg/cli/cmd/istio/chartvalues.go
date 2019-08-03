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

package istio

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/banzaicloud/backyards-cli/pkg/helm"
)

type Values struct {
	NameOverride         string `json:"nameOverride,omitempty"`
	FullnameOverride     string `json:"fullnameOverride,omitempty"`
	ReplicaCount         int    `json:"replicaCount"`
	UseNamespaceResource bool   `json:"useNamespaceResource"`

	RBAC struct {
		Enabled bool `json:"enabled"`
	} `json:"rbac"`

	PrometheusMetrics struct {
		Enabled   bool `json:"enabled"`
		AuthProxy struct {
			Enabled bool       `json:"enabled"`
			Image   helm.Image `json:"image"`
		} `json:"authProxy"`
	} `json:"prometheusMetrics"`

	Operator struct {
		Resources corev1.ResourceRequirements `json:"resources,omitempty"`
		Image     helm.Image                  `json:"image"`
	} `json:"operator"`

	IstioVersion string `json:"istioVersion"`
}

func (values *Values) SetDefaults(releaseName string) {
	values.NameOverride = releaseName
	values.UseNamespaceResource = true
}
