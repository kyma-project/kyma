package controllers

import (
	"fmt"
	"strings"

	serverless "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	//TODO make it type
	ConfigMapFunction = "handler.js"
	ConfigMapHandler  = "handler.main"
	ConfigMapDeps     = "package.json"
)

// creates config map containing js function
// and it's dependencies used during lambda image build
func configMap(fn *serverless.Function) *corev1.ConfigMap {
	data := map[string]string{
		ConfigMapHandler:  "handler.main",
		ConfigMapFunction: fn.Spec.Function,
		ConfigMapDeps:     "{}",
	}
	if strings.Trim(fn.Spec.Deps, " ") != "" {
		data["package.json"] = fn.Spec.Deps
	}

	cmLabels := make(map[string]string, len(fn.Labels)+1)
	for k, v := range fn.Labels {
		cmLabels[k] = v
	}
	cmLabels[serverless.FnUUID] = string(fn.UID)
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Labels:    cmLabels,
			Namespace: fn.Namespace,
			Name:      generateName(fmt.Sprintf("%s-", fn.Name)),
		},
		Data: data,
	}
}
