## Traffic shifting

### View traffic shifting rules

You can list the traffic shifting rules of a service in a given namespace with the following command:

```
$ backyards routing ts get backyards-demo/movies
INFO[0000] traffic shifting for backyards-demo/movies is currently set to v1=33, v2=33, v3=34
```

### Set traffic shifting rules

Now let's make the same changes through the CLI.

```
$ backyards routing ts set backyards-demo/movies v2=100
INFO[0001] traffic shifting for backyards-demo/movies set to v2=100 successfully
```

To verify that the command was successful:

```
$ backyards routing ts get backyards-demo/movies
INFO[0000] traffic shifting for backyards-demo/movies is currently set to v2=100
```

### Remove traffic shifting rules

To remove the traffic shifting rules:

```
$ backyards routing ts delete backyards-demo/movies
INFO[0001] traffic shifting rules set to backyards-demo/movies successfully deleted
```

To verify that the command was successful:

```
$ backyards routing ts get backyards-demo/movies
INFO[0000] no traffic shifting rules set for backyards-demo/movies
```
