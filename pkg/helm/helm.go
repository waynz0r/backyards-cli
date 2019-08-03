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
	"bytes"
	"net/http"
	"path"
	"strings"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/renderutil"
	"k8s.io/helm/pkg/timeconv"

	"istio.io/operator/pkg/object"
)

type ReleaseOptions chartutil.ReleaseOptions

func GetDefaultValues(fs http.FileSystem) ([]byte, error) {
	file, err := fs.Open(chartutil.ValuesfileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(file)

	return buf.Bytes(), nil
}

func Render(fs http.FileSystem, values string, releaseOptions ReleaseOptions) (object.K8sObjects, error) {
	chrtConfig := &chart.Config{
		Raw:    values,
		Values: map[string]*chart.Value{},
	}

	files := []*chartutil.BufferedFile{
		{
			Name: chartutil.ChartfileName,
		},
	}

	dir, err := fs.Open(chartutil.TemplatesDir)
	if err != nil {
		return nil, err
	}

	chartFiles, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}

	for _, chartFile := range chartFiles {
		filename := chartFile.Name()
		if strings.HasSuffix(filename, "yaml") || strings.HasSuffix(filename, "yml") || strings.HasSuffix(filename, "tpl") {
			files = append(files, &chartutil.BufferedFile{
				Name: chartutil.TemplatesDir + "/" + filename,
			})
		}
	}

	for _, f := range files {
		data, err := readIntoBytes(fs, f.Name)
		if err != nil {
			return nil, err
		}

		data = append(data, []byte("\n---\n")...)

		f.Data = data
	}

	// Create chart and render templates
	chrt, err := chartutil.LoadFiles(files)
	if err != nil {
		return nil, err
	}

	renderOpts := renderutil.Options{
		ReleaseOptions: chartutil.ReleaseOptions{
			Name:      releaseOptions.Name,
			IsInstall: true,
			IsUpgrade: false,
			Time:      timeconv.Now(),
			Namespace: releaseOptions.Namespace,
		},
		KubeVersion: "",
	}

	renderedTemplates, err := renderutil.Render(chrt, chrtConfig, renderOpts)
	if err != nil {
		return nil, err
	}

	// Merge templates and inject
	var buf bytes.Buffer
	for _, tmpl := range files {
		t := path.Join(renderOpts.ReleaseOptions.Name, tmpl.Name)
		if _, err := buf.WriteString(renderedTemplates[t]); err != nil {
			return nil, err
		}
	}

	objects, err := object.ParseK8sObjectsFromYAMLManifest(buf.String())
	if err != nil {
		return nil, err
	}

	return objects, nil
}

func readIntoBytes(fs http.FileSystem, filename string) ([]byte, error) {
	file, err := fs.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(file)

	return buf.Bytes(), nil
}

func InstallObjectOrder() func(o *object.K8sObject) int {
	var Order = []string{
		"CustomResourceDefinition",
		"Namespace",
		"ResourceQuota",
		"LimitRange",
		"PodSecurityPolicy",
		"PodDisruptionBudget",
		"Secret",
		"ConfigMap",
		"StorageClass",
		"PersistentVolume",
		"PersistentVolumeClaim",
		"ServiceAccount",
		"ClusterRole",
		"ClusterRoleList",
		"ClusterRoleBinding",
		"ClusterRoleBindingList",
		"Role",
		"RoleList",
		"RoleBinding",
		"RoleBindingList",
		"Service",
		"DaemonSet",
		"Pod",
		"ReplicationController",
		"ReplicaSet",
		"Deployment",
		"HorizontalPodAutoscaler",
		"StatefulSet",
		"Job",
		"CronJob",
		"Ingress",
		"APIService",
	}

	order := make(map[string]int, len(Order))
	for i, kind := range Order {
		order[kind] = i
	}

	return func(o *object.K8sObject) int {
		if nr, ok := order[o.Kind]; ok {
			return nr
		}
		return 1000
	}
}

func UninstallObjectOrder() func(o *object.K8sObject) int {
	var Order = []string{
		"APIService",
		"Ingress",
		"Service",
		"CronJob",
		"Job",
		"StatefulSet",
		"HorizontalPodAutoscaler",
		"Deployment",
		"ReplicaSet",
		"ReplicationController",
		"Pod",
		"DaemonSet",
		"RoleBindingList",
		"RoleBinding",
		"RoleList",
		"Role",
		"ClusterRoleBindingList",
		"ClusterRoleBinding",
		"ClusterRoleList",
		"ClusterRole",
		"ServiceAccount",
		"PersistentVolumeClaim",
		"PersistentVolume",
		"StorageClass",
		"ConfigMap",
		"Secret",
		"PodDisruptionBudget",
		"PodSecurityPolicy",
		"LimitRange",
		"ResourceQuota",
		"Namespace",
		"CustomResourceDefinition",
	}

	order := make(map[string]int, len(Order))
	for i, kind := range Order {
		order[kind] = i
	}

	return func(o *object.K8sObject) int {
		if nr, ok := order[o.Kind]; ok {
			return nr
		}
		return 1000
	}
}
