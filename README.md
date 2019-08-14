This is a command line interface under heavy development for [Backyards](https://banzaicloud.com/blog/istio-the-easy-way/).

### Installation

Pre-built binaries are available in multiple package formats. Download the [latest release](https://github.com/banzaicloud/backyards-cli/releases).

### Use

```text
Install and manage Backyards

Usage:
  backyards [command]

Available Commands:
  canary      Install and manage Canary feature
  dashboard   Open the Backyards dashboard in a web browser
  demoapp     Install and manage demo application
  help        Help about any command
  install     Install Backyards
  istio       Install and manage Istio
  uninstall   Uninstall Backyards
  version     Print the client and api version information

Flags:
      --context string      Name of the kubeconfig context to use
  -h, --help                Help for backyards
  -c, --kubeconfig string   Path to the kubeconfig file to use for CLI requests
  -n, --namespace string    Namespace in which Backyards is installed [$BACKYARDS_NAMESPACE] (default "backyards-system")
  -v, --verbose             Turn on debug logging
      --version             Version for backyards

Use "backyards [command] --help" for more information about a command.
```
