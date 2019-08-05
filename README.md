This is a command line interface under heavy development for [Backyards](https://banzaicloud.com/blog/istio-the-easy-way/).

### Installation

```bash
❯ go get github.com/banzaicloud/backyards-cli/cmd/backyards
```

### Use

```text
install and manage Backyards

Usage:
  backyards [command]

Available Commands:
  canary      install and manage Canary feature
  dashboard   Open the Backyards dashboard in a web browser
  demoapp     install and manage demo application
  help        Help about any command
  install     Install Backyards
  istio       install and manage Istio
  uninstall   Uninstalls Backyards
  version     Print the client and api version information

Flags:
      --context string      Name of the kubeconfig context to use
  -h, --help                help for backyards
  -c, --kubeconfig string   Path to the kubeconfig file to use for CLI requests
  -n, --namespace string    Namespace in which Backyards is installed [$BACKYARDS_NAMESPACE] (default "backyards-system")
  -v, --verbose             Turn on debug logging
      --version             version for backyards

Use "backyards [command] --help" for more information about a command.
```
