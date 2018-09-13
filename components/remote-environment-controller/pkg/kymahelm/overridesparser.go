package kymahelm

import (
	"bytes"
	"text/template"
)

func ParseOverrides(data interface{}, rawTemplate string) (string, error) {
	tmpl, err := template.New("").Parse(rawTemplate)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, data)
	if err != nil {
		return "", err
	}

	return string(buf.Bytes()), nil
}
