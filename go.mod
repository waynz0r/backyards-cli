module github.com/banzaicloud/backyards-cli

go 1.12

require (
	emperror.dev/errors v0.4.2
	emperror.dev/handler/logrus v0.1.0
	github.com/MakeNowJust/heredoc v0.0.0-20171113091838-e9091a26100e
	github.com/Masterminds/sprig v2.20.0+incompatible // indirect
	github.com/banzaicloud/istio-operator v0.0.0-20190821151858-a47cd7d9bc7a
	github.com/banzaicloud/k8s-objectmatcher v1.0.0
	github.com/ghodss/yaml v1.0.0
	github.com/machinebox/graphql v0.2.2
	github.com/matryer/is v1.2.0 // indirect
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.3
	github.com/spf13/viper v1.4.0
	go.uber.org/multierr v1.1.0
	istio.io/operator v0.0.0-20190805193245-ce3cfb6e2672
	k8s.io/api v0.0.0-20190620084959-7cf5895f2711
	k8s.io/apiextensions-apiserver v0.0.0-20190426053235-842c4571cde0
	k8s.io/apimachinery v0.0.0-20190612205821-1799e75a0719
	k8s.io/client-go v11.0.1-0.20190516230509-ae8359b20417+incompatible
	k8s.io/helm v2.14.3+incompatible
	knative.dev/pkg v0.0.0-20190903162800-3dd5d66573f6
	sigs.k8s.io/controller-runtime v0.2.0-beta.4
	sigs.k8s.io/yaml v1.1.0
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go => k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
)
