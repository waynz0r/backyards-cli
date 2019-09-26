## Circuit Breaking

### Set circuit breaking configurations

You can do this in interactive mode:

```
$ backyards r cb set backyards-demo/notifications
? Maximum number of HTTP1/TCP connections 1
? TCP connection timeout 3s
? Maximum number of pending HTTP requests 1
? Maximum number of requests 1024
? Maximum number of requests per connection 1
? Maximum number of retries 1024
? Number of errors before a host is ejected 1
? Time interval between ejection sweep analysis 1s
? Minimum ejection duration 3m
? Maximum ejection percentage 100
INFO[0043] circuit breaker rules successfully applied to 'backyards-demo/notifications'
Connections  Timeout  Pending Requests  Requests  RPC  Retries  Errors  Interval  Ejection time  percentage
1            3s       1                 1024      1    1024     1       1s        3m             100
```

Or, alternatively, in a non-interactive mode, by explicitly setting the values:

```
$ backyards r cb set backyards-demo/notifications --non-interactive --max-connections=1 --max-pending-requests=1 --max-requests-per-connection=1 --consecutiveErrors=1 --interval=1s --baseEjectionTime=3m --maxEjectionPercent=100
Connections  Timeout  Pending Requests  Requests  RPC  Retries  Errors  Interval  Ejection time  percentage
1            3s       1                 1024      1    1024     5       1s        3m             100
```

After the command is issued, the circuit breaking settings are fetched and displayed right away.

### View circuit breaking configurations

You can list the circuit breaking configurations of a service in a given namespace with the following command:

```
$ backyards r cb get backyards-demo/notifications
  Connections  Timeout  Pending Requests  Requests  RPC  Retries  Errors  Interval  Ejection time  percentage
  1            3s       1                 1024      1    1024     5       1s        3m             100
```

By default, the results are displayed in a table view, but it's also possible to list the configurations in `JSON` or `YAML` format:

```
$ backyards r cb get backyards-demo/notifications -o json
  {
    "maxConnections": 1,
    "connectTimeout": "3s",
    "http1MaxPendingRequests": 1,
    "http2MaxRequests": 1024,
    "maxRequestsPerConnection": 1,
    "maxRetries": 1024,
    "consecutiveErrors": 5,
    "interval": "1s",
    "baseEjectionTime": "3m",
    "maxEjectionPercent": 100
  }

$ backyards r cb get backyards-demo/notifications -o yaml
  maxConnections: 1
  connectTimeout: 3s
  http1MaxPendingRequests: 1
  http2MaxRequests: 1024
  maxRequestsPerConnection: 1
  maxRetries: 1024
  consecutiveErrors: 5
  interval: 1s
  baseEjectionTime: 3m
  maxEjectionPercent: 100
```

###  Monitor circuit breaking

To see similar dashboards from the CLI that can be seen on the Grafana dashboards on the UI as well, trigger circuit breaker trips by calling the service from multiple connections and then issue the following command:

```
$ backyards r cb graph backyards-demo/notifications
```

You should see something like this:

![Circuit Breaking trip cli](/docs/img/circuit-breaking-trip-cli.png)

### Remove circuit breaking configurations

To remove circuit breaking configurations:

```
$ backyards r cb delete backyards-demo/notifications
INFO[0000] current settings
Connections  Timeout  Pending Requests  Requests  RPC  Retries  Errors  Interval  Ejection time  percentage
1            3s       1                 1024      1    1024     5       1s        3m             100
? Do you want to DELETE the circuit breaker rules? Yes
INFO[0008] circuit breaker rules set to backyards-demo/notifications successfully deleted
```

To verify that the command was successful:

```
$ backyards r cb get backyards-demo/notifications
  INFO[0001] no circuit breaker rules set for backyards-demo/notifications
```
