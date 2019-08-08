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
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"istio.io/operator/pkg/object"
	appsv1 "k8s.io/api/apps/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	k8sclient "github.com/banzaicloud/backyards-cli/pkg/k8s/client"
)

type NamespacedNameWithGVK struct {
	types.NamespacedName
	schema.GroupVersionKind
}

type ResourceConditionCheck func(*unstructured.Unstructured, error) bool

func ExistsConditionCheck(obj *unstructured.Unstructured, k8serror error) bool {
	if k8serror == nil {
		return true
	}

	if !k8serrors.IsNotFound(k8serror) {
		log.Error(k8serror)
	}

	return false
}

func NonExistsConditionCheck(obj *unstructured.Unstructured, k8serror error) bool {
	if k8serrors.IsNotFound(k8serror) {
		return true
	}

	return false
}

func CRDEstablishedConditionCheck(obj *unstructured.Unstructured, k8serror error) bool {
	var resource apiextensionsv1beta1.CustomResourceDefinition
	err := k8sclient.GetScheme().Convert(obj, &resource, nil)
	// simply return true for unconvertable objects
	if err != nil {
		return true
	}

	for _, condition := range resource.Status.Conditions {
		if condition.Type == apiextensionsv1beta1.Established {
			if condition.Status == apiextensionsv1beta1.ConditionTrue {
				return true
			}
		}
	}

	return false
}

func ReadyReplicasConditionCheck(obj *unstructured.Unstructured, k8serror error) bool {
	var deployment appsv1.Deployment
	deploymentErr := k8sclient.GetScheme().Convert(obj, &deployment, nil)
	if deploymentErr == nil && deployment.Status.ReadyReplicas == deployment.Status.Replicas {
		return true
	}

	var statefulset appsv1.StatefulSet
	statefulsetErr := k8sclient.GetScheme().Convert(obj, &statefulset, nil)
	if statefulsetErr == nil && statefulset.Status.ReadyReplicas == statefulset.Status.Replicas {
		return true
	}

	// return true for unconvertable objects
	if deploymentErr != nil && statefulsetErr != nil {
		return true
	}

	return false
}

func WaitForResourcesConditions(client k8sclient.Client, objects []NamespacedNameWithGVK, backoff wait.Backoff, checkFuncs ...ResourceConditionCheck) error {
	for _, o := range objects {
		err := waitForResourceConditions(client, o.Unstructured(), backoff, checkFuncs...)
		if err != nil {
			return err
		}
	}
	return nil
}

type WaitForResourceConditionsFunc func(k8sclient.Client, *unstructured.Unstructured) error

func WaitForResourceConditions(backoff wait.Backoff, checkFuncs ...ResourceConditionCheck) WaitForResourceConditionsFunc {
	return func(client k8sclient.Client, object *unstructured.Unstructured) error {
		return waitForResourceConditions(client, object, backoff, checkFuncs...)
	}
}

func waitForResourceConditions(client k8sclient.Client, object *unstructured.Unstructured, backoff wait.Backoff, checkFuncs ...ResourceConditionCheck) error {
	log.Infof("%s - pending", getFormattedName(object))
	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		resource := object.DeepCopy()
		err := client.Get(context.Background(), types.NamespacedName{
			Name:      resource.GetName(),
			Namespace: resource.GetNamespace(),
		}, resource)
		for _, fn := range checkFuncs {
			ok := fn(resource, err)
			if !ok {
				return false, nil
			}
		}
		return true, nil
	})

	if err != nil {
		return err
	}

	log.Infof("%s - ok", getFormattedName(object))

	return nil
}

func (o NamespacedNameWithGVK) Unstructured() *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetName(o.Name)
	obj.SetNamespace(o.Namespace)
	obj.SetGroupVersionKind(o.GroupVersionKind)

	return obj
}

func (o NamespacedNameWithGVK) String() string {
	if o.Namespace == "" {
		return fmt.Sprintf("%s/%s", strings.ToLower(o.GroupVersionKind.Kind), o.Name)
	}
	return fmt.Sprintf("%s/%s/%s", strings.ToLower(o.GroupVersionKind.Kind), o.Namespace, o.Name)
}

func NamesWithGVKFromK8sObjects(objects object.K8sObjects, kind ...string) []NamespacedNameWithGVK {
	kinds := make(map[string]bool, len(kind))
	for _, kind := range kind {
		kinds[kind] = true
	}
	names := make([]NamespacedNameWithGVK, 0)
	for _, obj := range objects {
		if len(kinds) > 0 && !kinds[obj.GroupVersionKind().Kind] {
			continue
		}
		names = append(names, NamespacedNameWithGVK{
			NamespacedName: types.NamespacedName{
				Name:      obj.Name,
				Namespace: obj.Namespace,
			},
			GroupVersionKind: obj.GroupVersionKind(),
		})
	}

	return names
}
