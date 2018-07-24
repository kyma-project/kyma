package notifier

const (
	header = "_Tests execution summary report from last *{{ .TestResultWindowTime }}*_"
	body   = `
	{{- if .FailedTests }}
	*Summary:* {{ .TotalTestsCnt }} test executions and {{ len .FailedTests }} of them failed :sad-frog:
		{{block "list" .FailedTests }}
		Failed tests IDs:
			{{ range . }}
			{{printf "- %q" .ID }}
			{{- end}}
		{{end}}
	{{- else -}}
	*Summary:* {{ .TotalTestsCnt }} test executions and all of them passed :very_nice:
	{{- end -}}`
	footer = `
	{{- if .FailedTests }}
	_*Run*_` +
		"```" +
		"kubectl exec -n {{ .TestRunnerInfo.Namespace }} {{ .TestRunnerInfo.PodName }} -- logs-printer --ids=" + `
			{{- range $index, $element := .FailedTests -}}
					{{- if ne $index 0 -}},{{- end -}}
					{{- $element.ID -}}
			{{- end -}}` +
		"```" +
		` _*to get more info about failed tests.*_
	{{- end -}}
`
)
