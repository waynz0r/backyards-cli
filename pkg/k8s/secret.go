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

package k8s

import (
	"context"

	"emperror.dev/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	k8sclient "github.com/banzaicloud/backyards-cli/pkg/k8s/client"
)

// GetTokenForServiceAccountName retrieves an auth token of service account from the related secret
func GetTokenForServiceAccountName(client k8sclient.Client, key types.NamespacedName) (string, error) {
	var sa corev1.ServiceAccount
	var secret corev1.Secret
	var token string

	err := client.Get(context.Background(), key, &sa)
	if err != nil {
		return token, errors.WrapIfWithDetails(err, "could not get service account", "name", key.String())
	}

	if len(sa.Secrets) < 1 {
		return token, errors.NewWithDetails("no secrets found in service account", "name", key.String())
	}

	secretKey := types.NamespacedName{
		Name:      sa.Secrets[0].Name,
		Namespace: key.Namespace,
	}
	err = client.Get(context.Background(), secretKey, &secret)
	if err != nil {
		return token, errors.WrapIfWithDetails(err, "could not get secret", "name", secretKey.String())
	}

	token = string(secret.Data[corev1.ServiceAccountTokenKey])

	return token, nil
}
