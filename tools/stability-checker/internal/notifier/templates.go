package notifier

const (
	header = `_Tests execution summary report from last *{{ .TestResultWindowTime }}*_
	{{- if .TestRunnerInfo.ClusterName }}{{"\n"}}_Running on the *{{ .TestRunnerInfo.ClusterName }}* cluster_{{- end }}`
	body = `
	{{- if .FailedExecutions }}
	*Summary:* {{ .TotalTestsCnt }} test executions and {{ len .FailedExecutions }} of them failed :sad-frog:
	{{- if .ShowTestStats -}}
 	  {{"\n"}}Test statistics from failed executions:
      {{- range .TestStats -}}
        {{"\n   "}} {{.Failures}} failures of {{ .Name }}
	  {{- else }}
		No test statistics
	  {{- end }}
	{{- end }}

		{{block "list" .FailedExecutions }}
		Failed test executions IDs:
			{{ range . }}
			{{printf "- %q" .ID }}
			{{- end}}
		{{end}}
	{{- else -}}
	*Summary:* {{ .TotalTestsCnt }} test executions and all of them passed :very_nice:
	{{- end -}}
	`
	code   = "```"
	footer = `
	{{- if .FailedExecutions }}
	_*Run*_` +
		code +
		`kubectl logs -n {{ .TestRunnerInfo.Namespace }} {{ .TestRunnerInfo.PodName }} stability-checker | jq -s '.[]| select(
			{{- range $index, $element := .FailedExecutions -}}
					{{- if ne $index 0 }} or {{ end -}}
					.log."test-run-id" == "{{- $element.ID -}}"
			{{- end -}}
		)' | jq -r '.log.time + "\n" + .log.message'` +
		code +
		` _*to get more info about failed tests.*_
	{{- end -}}
`
)
