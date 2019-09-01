# Backyards Helm Chart

## TL;DR

```bash
$ helm repo add banzaicloud-stable https://kubernetes-charts.banzaicloud.com
$ helm repo update
$ helm install --name=backyards banzaicloud-stable/backyards
```

## Introduction

[Backyards](https://github.com/banzaicloud/backyards) manages hybrid clusters connected with [Istio](https://istio.io/).

## Prerequisites

- Kubernetes 1.10+

## Installing the Chart

To install the chart with the release name `backyards`:

```bash
helm install --name backyards banzaicloud-stable/backyards
```

The command deploys the application on the Kubernetes cluster with the default configuration.
The configuration section lists the parameters that can be configured during installation.

> Tip: List all releases using `helm list`

## Uninstalling the Chart

To uninstall/delete the backyards release:

```bash
$ helm del --purge backyards
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The configurable parameters and default values are listed in [`values.yaml`](values.yaml).

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.

Alternatively, a YAML file that specifies the values for the parameters can be provided during the chart installation:

```bash
$ helm install --name backyards -f my-values.yaml banzaicloud-stable/backyards
```
