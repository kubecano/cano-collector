{{/*
Expand the name of the chart.
*/}}
{{- define "cano-collector.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "cano-collector.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "cano-collector.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "cano-collector.labels" -}}
helm.sh/chart: {{ include "cano-collector.chart" . }}
{{ include "cano-collector.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "cano-collector.selectorLabels" -}}
app.kubernetes.io/name: {{ include "cano-collector.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Validate Slack destination configuration
*/}}
{{- define "cano-collector.validateSlackDestination" -}}
{{- $destination := . -}}
{{- $hasApiKey := hasKey $destination "api_key" -}}
{{- $hasApiKeyValueFrom := hasKey $destination "api_key_value_from" -}}
{{- $authMethods := 0 -}}
{{- if $hasApiKey }}{{ $authMethods = add $authMethods 1 }}{{ end -}}
{{- if $hasApiKeyValueFrom }}{{ $authMethods = add $authMethods 1 }}{{ end -}}
{{- if eq $authMethods 0 -}}
{{- fail (printf "Slack destination '%s' must have exactly one of: api_key or api_key_value_from" $destination.name) -}}
{{- else if gt $authMethods 1 -}}
{{- fail (printf "Slack destination '%s' cannot have multiple authentication methods. Use only one of: api_key or api_key_value_from" $destination.name) -}}
{{- end -}}
{{- if $hasApiKeyValueFrom -}}
{{- if not (hasKey $destination.api_key_value_from "secretName") -}}
{{- fail (printf "Slack destination '%s' api_key_value_from must have secretName" $destination.name) -}}
{{- end -}}
{{- if not (hasKey $destination.api_key_value_from "secretKey") -}}
{{- fail (printf "Slack destination '%s' api_key_value_from must have secretKey" $destination.name) -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Resolve API key from external secret
*/}}
{{- define "cano-collector.resolveApiKey" -}}
{{- $destination := .destination -}}
{{- $context := .context -}}
{{- if hasKey $destination "api_key_value_from" -}}
{{- $secretName := $destination.api_key_value_from.secretName -}}
{{- $secretKey := $destination.api_key_value_from.secretKey -}}
{{- $secret := (lookup "v1" "Secret" $context.Release.Namespace $secretName) -}}
{{- if not $secret -}}
{{- fail (printf "Secret '%s' not found in namespace '%s' for Slack destination '%s'" $secretName $context.Release.Namespace $destination.name) -}}
{{- end -}}
{{- $apiKey := (index $secret.data $secretKey) -}}
{{- if not $apiKey -}}
{{- fail (printf "Key '%s' not found in secret '%s' for Slack destination '%s'" $secretKey $secretName $destination.name) -}}
{{- end -}}
{{- $apiKey | b64dec | quote -}}
{{- else if hasKey $destination "api_key" -}}
{{- $destination.api_key | quote -}}
{{- end -}}
{{- end -}}
