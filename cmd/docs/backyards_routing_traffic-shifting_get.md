## backyards routing traffic-shifting get

Get traffic shifting rules for a service

### Synopsis

Get traffic shifting rules for a service

```
backyards routing traffic-shifting get [[--service=]namespace/servicename] [flags]
```

### Options

```
  -h, --help             help for get
      --service string   Service name
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

* [backyards routing traffic-shifting](backyards_routing_traffic-shifting.md)	 - Manage traffic-shifting configurations

###### Auto generated by spf13/cobra on 26-Sep-2019
