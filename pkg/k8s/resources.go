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

	"emperror.dev/errors"
	log "github.com/sirupsen/logrus"
	"istio.io/operator/pkg/object"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	k8sclient "github.com/banzaicloud/backyards-cli/pkg/k8s/client"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
)

type Object interface {
	metav1.Object
	metav1.Type
	schema.ObjectKind
}

type PostResourceApplyFunc func(k8sclient.Client, Object) error

func ApplyResources(client k8sclient.Client, objects object.K8sObjects, postResourceApplyFuncs ...PostResourceApplyFunc) error {
	var err error

	for _, obj := range objects {
		actual := obj.UnstructuredObject().DeepCopy()
		desired := obj.UnstructuredObject().DeepCopy()

		var group string
		if desired.GroupVersionKind().Group != "" {
			group = "." + desired.GroupVersionKind().Group
		}
		objectName := fmt.Sprintf("%s%s/%s", strings.ToLower(desired.GetKind()), group, desired.GetName())

		if err = client.Get(context.Background(), types.NamespacedName{
			Name:      actual.GetName(),
			Namespace: actual.GetNamespace(),
		}, actual); err == nil {
			desired.SetResourceVersion(actual.GetResourceVersion())
			patchResult, err := patch.DefaultPatchMaker.Calculate(actual, desired)
			if err != nil {
				log.Error(err, "could not match objects", "object", actual.GetKind())
			} else if patchResult.IsEmpty() {
				log.Infof("%s unchanged", objectName)
				continue
			}

			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(desired); err != nil {
				log.Error(err, "failed to set last applied annotation", "desired", desired)
			}

			err = client.Update(context.Background(), desired)
			if err != nil {
				return errors.WrapIfWithDetails(err, "could not update resource", "name", objectName)
			}
			log.Infof("%s configured", objectName)
		} else {
			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(desired); err != nil {
				log.Error(err, "failed to set last applied annotation", "desired", desired)
			}

			err = client.Create(context.Background(), desired)
			if err != nil {
				return errors.WrapIfWithDetails(err, "could not create resource", "name", objectName)
			}
			log.Infof("%s created", objectName)
		}

		if len(postResourceApplyFuncs) > 0 {
			for _, fn := range postResourceApplyFuncs {
				err = fn(client, actual)
				if err != nil {
					log.Error(err)
				}
			}
		}
	}

	return nil
}

type PostResourceDeleteFunc func(k8sclient.Client, Object) error

func DeleteResources(client k8sclient.Client, objects object.K8sObjects, postResourceDeleteFuncs ...PostResourceDeleteFunc) error {
	var err error

	for _, obj := range objects {
		actual := obj.UnstructuredObject().DeepCopy()
		objectName := getFormattedName(actual)
		if err = client.Get(context.Background(), types.NamespacedName{
			Name:      actual.GetName(),
			Namespace: actual.GetNamespace(),
		}, actual); err == nil {
			err = client.Delete(context.Background(), obj.UnstructuredObject())
			if k8serrors.IsNotFound(err) || k8smeta.IsNoMatchError(err) {
				log.Error(errors.WrapIf(err, "could not delete"))
				continue
			}
			if err != nil {
				log.Error(err)
			}

			if len(postResourceDeleteFuncs) > 0 {
				for _, fn := range postResourceDeleteFuncs {
					err = fn(client, actual)
					if err != nil {
						log.Error(err)
						continue
					}
				}
			}

			log.Infof("%s deleted", objectName)
		} else {
			err = errors.WrapIf(err, "could not delete")
			if k8serrors.IsNotFound(err) {
				log.Warning(err)
			} else {
				log.Error(err)
			}
		}
	}

	return nil
}

func WaitForCRD(backoff wait.Backoff) PostResourceApplyFunc {
	return func(client k8sclient.Client, resource Object) error {
		if resource.GetKind() != "CustomResourceDefinition" {
			return nil
		}

		objectName := getFormattedName(resource)

		wait.ExponentialBackoff(backoff, func() (bool, error) {
			var crd apiextensionsv1beta1.CustomResourceDefinition
			log.Debugf("wait for %s to be available", objectName)
			err := client.Get(context.Background(), types.NamespacedName{
				Name:      resource.GetName(),
				Namespace: resource.GetNamespace(),
			}, &crd)
			if err == nil {
				for _, cond := range crd.Status.Conditions {
					switch cond.Type {
					case apiextensionsv1beta1.Established:
						if cond.Status == apiextensionsv1beta1.ConditionTrue {
							return true, nil
						}
					case apiextensionsv1beta1.NamesAccepted:
						if cond.Status == apiextensionsv1beta1.ConditionFalse {
							return false, errors.New(cond.Reason)
						}
					}
				}
			} else {
				log.Error(err)
			}
			return false, nil
		})

		return nil
	}
}

func WaitForFinalizers(backoff wait.Backoff) PostResourceDeleteFunc {
	return func(client k8sclient.Client, resource Object) error {
		if len(resource.GetFinalizers()) > 0 {
			objectName := getFormattedName(resource)
			err := wait.ExponentialBackoff(backoff, func() (bool, error) {
				obj := resource.(*unstructured.Unstructured)
				log.Debugf("wait for %s to be deleted", objectName)
				err := client.Get(context.Background(), types.NamespacedName{
					Name:      resource.GetName(),
					Namespace: resource.GetNamespace(),
				}, obj)
				if k8serrors.IsNotFound(err) {
					return true, nil
				}
				return false, nil
			})
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func getFormattedName(object Object) string {
	var group string
	if object.GroupVersionKind().Group != "" {
		group = "." + object.GroupVersionKind().Group
	}

	return fmt.Sprintf("%s%s/%s", strings.ToLower(object.GetKind()), group, object.GetName())
}
