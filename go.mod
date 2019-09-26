module github.com/banzaicloud/backyards-cli

go 1.13

require (
	emperror.dev/errors v0.4.2
	emperror.dev/handler/logrus v0.1.0
	github.com/AlecAivazis/survey/v2 v2.0.2
	github.com/MakeNowJust/heredoc v0.0.0-20171113091838-e9091a26100e
	github.com/Masterminds/sprig v2.20.0+incompatible // indirect
	github.com/banzaicloud/istio-operator v0.0.0-20190821151858-a47cd7d9bc7a
	github.com/banzaicloud/k8s-objectmatcher v1.0.1
	github.com/machinebox/graphql v0.2.2
	github.com/matryer/is v1.2.0 // indirect
	github.com/mattn/go-isatty v0.0.8
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	github.com/ttacon/chalk v0.0.0-20160626202418-22c06c80ed31
	github.com/waynz0r/grafterm v0.2.1-0.20190814214739-b7722452f1e4
	go.uber.org/multierr v1.1.0
	gomodules.xyz/jsonpatch/v2 v2.0.1 // indirect
	gopkg.in/yaml.v2 v2.2.2
	istio.io/operator v0.0.0-20190805193245-ce3cfb6e2672
	k8s.io/api v0.0.0-20190704095032-f4ca3d3bdf1d
	k8s.io/apiextensions-apiserver v0.0.0-20190801143813-8b5f3a974f92
	k8s.io/apimachinery v0.0.0-20190704094733-8f6ac2502e51
	k8s.io/client-go v11.0.1-0.20190516230509-ae8359b20417+incompatible
	k8s.io/helm v2.14.3+incompatible
	knative.dev/pkg v0.0.0-20190903162800-3dd5d66573f6
	sigs.k8s.io/controller-runtime v0.2.0-beta.4
	sigs.k8s.io/yaml v1.1.0
)
