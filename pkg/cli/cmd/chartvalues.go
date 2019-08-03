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

package cmd

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/banzaicloud/backyards-cli/pkg/helm"
)

type Values struct {
	NameOverride         string                      `json:"nameOverride,omitempty"`
	FullnameOverride     string                      `json:"fullnameOverride,omitempty"`
	ReplicaCount         int                         `json:"replicaCount"`
	UseNamespaceResource bool                        `json:"useNamespaceResource"`
	Resources            corev1.ResourceRequirements `json:"resources,omitempty"`

	Ingress struct {
		Enabled     bool              `json:"enabled"`
		Annotations map[string]string `json:"annotations"`
		Paths       struct {
			Application string `json:"application"`
			Web         string `json:"web"`
		} `json:"paths"`
		BasePath string   `json:"basePath"`
		Hosts    []string `json:"hosts"`
		TLS      []struct {
			SecretName string   `json:"secretName"`
			Hosts      []string `json:"hosts"`
		} `json:"tls"`
	} `json:"ingress"`

	Autoscaling struct {
		Enabled                           bool `json:"enabled"`
		MinReplicas                       int  `json:"minReplicas"`
		MaxReplicas                       int  `json:"maxReplicas"`
		TargetCPUUtilizationPercentage    int  `json:"targetCPUUtilizationPercentage"`
		TargetMemoryUtilizationPercentage int  `json:"targetMemoryUtilizationPercentage"`
	} `json:"autoscaling"`

	Application struct {
		helm.EnvironmentVariables
		Image   helm.Image `json:"image"`
		Service struct {
			Type string `json:"type"`
			Port int    `json:"port"`
		} `json:"service"`
	} `json:"application"`

	Web struct {
		helm.EnvironmentVariables
		Enabled   bool                        `json:"enabled"`
		Image     helm.Image                  `json:"image"`
		Resources corev1.ResourceRequirements `json:"resources,omitempty"`
		Service   struct {
			Type string `json:"type"`
			Port int    `json:"port"`
		} `json:"service"`
	} `json:"web"`

	Istio struct {
		Namespace          string `json:"namespace"`
		CRName             string `json:"CRName"`
		ServiceAccountName string `json:"serviceAccountName"`
	} `json:"istio"`

	Prometheus struct {
		Enabled     bool                        `json:"enabled"`
		Image       helm.Image                  `json:"image"`
		Resources   corev1.ResourceRequirements `json:"resources,omitempty"`
		ExternalURL string                      `json:"externalUrl"`
		Config      struct {
			Global struct {
				ScrapeInterval     string `json:"scrapeInterval"`
				ScrapeTimeout      string `json:"scrapeTimeout"`
				EvaluationInterval string `json:"evaluationInterval"`
			} `json:"global"`
		} `json:"config"`
		Service struct {
			Type string `json:"type"`
			Port int    `json:"port"`
		} `json:"service"`
	} `json:"prometheus"`

	IngressGateway struct {
		Service struct {
			Type string `json:"type"`
		} `json:"service"`
	} `json:"ingressgateway"`
}

func (values *Values) SetDefaults(releaseName, istioNamespace string) {
	values.NameOverride = releaseName
	values.UseNamespaceResource = true
	values.Resources = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("128Mi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("200m"),
			corev1.ResourceMemory: resource.MustParse("256Mi"),
		},
	}

	values.Ingress.Enabled = false

	values.Web.Enabled = true
	values.Web.Resources = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("128Mi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("200m"),
			corev1.ResourceMemory: resource.MustParse("256Mi"),
		},
	}

	values.Prometheus.Enabled = true
	values.Prometheus.Resources = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("128Mi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("200m"),
			corev1.ResourceMemory: resource.MustParse("256Mi"),
		},
	}
	values.Prometheus.ExternalURL = "/prometheus"
	values.Prometheus.Config.Global.ScrapeInterval = "10s" //nolint
	values.Prometheus.Config.Global.ScrapeTimeout = "10s"
	values.Prometheus.Config.Global.EvaluationInterval = "10s"
}
