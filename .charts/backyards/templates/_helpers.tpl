{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "backyards.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "prometheus.name" -}}
{{- printf "%s-prometheus" (include "backyards.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "ingressgateway.name" -}}
{{- printf "%s-ingressgateway" (include "backyards.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "backyards.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create a default fully qualified app name for the Prometheus component
*/}}
{{- define "prometheus.fullname" -}}
{{- printf "%s-prometheus" (include "backyards.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name for the Ingress gateway component
*/}}
{{- define "ingressgateway.fullname" -}}
{{- printf "%s-ingressgateway" (include "backyards.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "backyards.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Call nested templates.
Source: https://stackoverflow.com/a/52024583/3027614
*/}}
{{- define "call-nested" }}
{{- $dot := index . 0 }}
{{- $subchart := index . 1 }}
{{- $template := index . 2 }}
{{- include $template (dict "Chart" (dict "Name" $subchart) "Values" (index $dot.Values $subchart) "Release" $dot.Release "Capabilities" $dot.Capabilities) }}
{{- end -}}
