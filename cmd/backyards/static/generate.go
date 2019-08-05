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
// +build ignore

package main

import (
	"github.com/shurcooL/vfsgen"
	log "github.com/sirupsen/logrus"

	"github.com/banzaicloud/backyards-cli/cmd/backyards/static"
)

func main() {
	err := vfsgen.Generate(static.BackyardsChartSource, vfsgen.Options{
		Filename:     "static/backyards/chart.gogen.go",
		PackageName:  "backyards",
		VariableName: "Chart",
	})
	if err != nil {
		log.Fatalln(err)
	}

	err = vfsgen.Generate(static.IstioOperatorChartSource, vfsgen.Options{
		Filename:     "static/istio_operator/chart.gogen.go",
		PackageName:  "istio_operator",
		VariableName: "Chart",
	})
	if err != nil {
		log.Fatalln(err)
	}

	err = vfsgen.Generate(static.CanaryOperatorChartSource, vfsgen.Options{
		Filename:     "static/canary_operator/chart.gogen.go",
		PackageName:  "canary_operator",
		VariableName: "Chart",
	})
	if err != nil {
		log.Fatalln(err)
	}

	err = vfsgen.Generate(static.MeshdemoChartSource, vfsgen.Options{
		Filename:     "static/meshdemo/chart.gogen.go",
		PackageName:  "meshdemo",
		VariableName: "Chart",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
