{{/*
Create a URL for container images, without version number
*/}}
{{- define "shortimageurl" -}}
{{- $registry := default $.reg.path $.img.containerRegistryPath -}}
{{- if hasKey $.img "directory" -}}
{{- printf "%s/%s/%s" $registry $.img.directory $.img.name -}}
{{- else -}}
{{- printf "%s/%s" $registry $.img.name -}}
{{- end -}}
{{- end -}}
