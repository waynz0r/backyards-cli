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
	err := vfsgen.Generate(static.BackyardsChart, vfsgen.Options{
		Filename:     "static/generated_backyards_chart.gogen.go",
		PackageName:  "static",
		BuildTags:    "prod",
		VariableName: "BackyardsChart",
	})
	if err != nil {
		log.Fatalln(err)
	}

	err = vfsgen.Generate(static.IstioOperatorChart, vfsgen.Options{
		Filename:     "static/generated_istio_operator_chart.gogen.go",
		PackageName:  "static",
		BuildTags:    "prod",
		VariableName: "IstioOperatorChart",
	})
	if err != nil {
		log.Fatalln(err)
	}

	err = vfsgen.Generate(static.CanaryOperatorChart, vfsgen.Options{
		Filename:     "static/generated_canary_operator_chart.gogen.go",
		PackageName:  "static",
		BuildTags:    "prod",
		VariableName: "CanaryOperatorChart",
	})
	if err != nil {
		log.Fatalln(err)
	}

	err = vfsgen.Generate(static.MeshdemoChart, vfsgen.Options{
		Filename:     "static/generated_meshdemo_chart.gogen.go",
		PackageName:  "static",
		BuildTags:    "prod",
		VariableName: "MeshdemoChart",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
