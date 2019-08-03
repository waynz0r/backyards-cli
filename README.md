This is a command line interface under heavy development for [Backyards](https://banzaicloud.com/blog/istio-the-easy-way/).

### Installation

Use the following command to quickly install the CLI:

```bash
❯ curl https://getbackyards.sh/cli | sh
```

The [script](scripts/getcli.sh) automatically chooses the best distribution package for your platform.

Available packages:

- [Debian package](https://banzaicloud.com/downloads/backyards-cli/latest?format=deb)
- [RPM package](https://banzaicloud.com/downloads/backyards-cli/latest?format=rpm)
- binary tarballs for [Linux](https://banzaicloud.com/downloads/backyards-cli/latest?os=linux) and [macOS](https://banzaicloud.com/downloads/backyards-cli/latest?os=darwin).

You can also select the installation method (one of `auto`, `deb`, `rpm`, `brew`, `tar` or `go`) explicitly:

```bash
❯ curl https://getbackyards.sh/cli | sh -s -- deb
```

Alternatively, fetch the source and compile it using `go get`:

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
