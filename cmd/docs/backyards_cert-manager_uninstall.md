## backyards cert-manager uninstall

Output or delete Kubernetes resources to uninstall cert-manager

### Synopsis

Output or delete Kubernetes resources to uninstall cert-manager.

The command automatically removes the resources.
It can only dump the removable resources with the '--dump-resources' option.

```
backyards cert-manager uninstall [flags]
```

### Examples

```
  # Default uninstall.
  backyards cert-manager uninstall

  # Uninstall cert-manager from a non-default namespace.
  backyards cert-manager uninstall --cert-manager-namespace backyards-system
```

### Options

```
  -d, --dump-resources   Dump resources to stdout instead of applying them
  -h, --help             help for uninstall
```

### Options inherited from parent commands

```
      --context string      name of the kubeconfig context to use
      --interactive         ask questions interactively even if stdin or stdout is non-tty
  -c, --kubeconfig string   path to the kubeconfig file to use for CLI requests
  -n, --namespace string    namespace in which Backyards is installed [$BACKYARDS_NAMESPACE] (default "backyards-system")
      --non-interactive     never ask questions interactively
  -o, --output string       output format (table|yaml|json) (default "table")
  -v, --verbose             turn on debug logging
```

### SEE ALSO

* [backyards cert-manager](backyards_cert-manager.md)	 - Install and manage cert-manager

###### Auto generated by spf13/cobra on 26-Sep-2019
