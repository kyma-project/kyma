package k8s

import (
	jsonencoder "encoding/json"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type json map[string]interface{}

func fixNamespace(name string) *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: name,
		},
	}
}

func stringifyJSON(in json) (string, error) {
	bytes, err := jsonencoder.Marshal(in)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
