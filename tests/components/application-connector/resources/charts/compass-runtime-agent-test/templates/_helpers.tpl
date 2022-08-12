
{{/*
Create a URL for container images
*/}}
{{- define "imageurl" -}}
{{- $registry := default $.reg.path $.img.containerRegistryPath -}}
{{- if hasKey $.img "directory" -}}
{{- printf "%s/%s/%s:%s" $registry $.img.directory $.img.name $.img.version -}}
{{- else -}}
{{- printf "%s/%s:%s" $registry $.img.name $.img.version -}}
{{- end -}}
{{- end -}}
