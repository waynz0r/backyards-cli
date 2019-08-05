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

// +build !prod

package static

import (
	"net/http"
	"path"
	"path/filepath"
	"runtime"
)

// Chart that will be rendered by `backyards install`
var BackyardsChartSource http.FileSystem = http.Dir(path.Join(getRepoRoot(), ".gen/charts/backyards"))

// Chart that will be rendered by `backyards istio install`
var IstioOperatorChartSource http.FileSystem = http.Dir(path.Join(getRepoRoot(), ".gen/charts/istio-operator"))

// Chart that will be rendered by `backyards canary install`
var CanaryOperatorChartSource http.FileSystem = http.Dir(path.Join(getRepoRoot(), ".gen/charts/canary-operator"))

// Chart that will be rendered by `backyards demoapp install`
var MeshdemoChartSource http.FileSystem = http.Dir(path.Join(getRepoRoot(), ".gen/charts/meshdemo"))

// getRepoRoot returns the full path to the root of the repo
func getRepoRoot() string {
	_, filename, _, _ := runtime.Caller(0)

	dir := filepath.Dir(filename)

	return filepath.Dir(path.Join(dir, "../.."))
}
