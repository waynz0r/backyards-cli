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

package helm

import (
	corev1 "k8s.io/api/core/v1"
)

type Image struct {
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	PullPolicy string `json:"pullPolicy"`
}

type Selectors struct {
	NodeSelector map[string]string   `json:"nodeSelector,omitempty"`
	Tolerations  []corev1.Toleration `json:"tolerations,omitempty"`
	Affinity     corev1.Affinity     `json:"affinity,omitempty"`
}

type EnvironmentVariables struct {
	Env        map[string]string `json:"env"`
	EnvSecrets []struct {
		Name         string                   `json:"name"`
		SecretKeyRef corev1.SecretKeySelector `json:"secretKeyRef"`
	} `json:"envSecrets"`
	EnvResourceField []struct {
		Name             string                       `json:"name"`
		ResourceFieldRef corev1.ResourceFieldSelector `json:"resourceFieldRef"`
	} `json:"envResourceField"`
	EnvConfigMap []struct {
		Name            string                      `json:"name"`
		ConfigMapKeyRef corev1.ConfigMapKeySelector `json:"configMapKeyRef"`
	} `json:"envConfigMaps"`
}
