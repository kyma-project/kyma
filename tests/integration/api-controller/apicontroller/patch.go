package apicontroller

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/types"

	kymaApi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	kyma "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
)

// A few utility functions to patch Api resources.
//
// Patch uses JSONPatch. There are some functions for specifying which operations should be performed.
// For the reference of other possible operations use https://tools.ietf.org/html/rfc6902#section-4.

type operation interface{}

type replaceOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

type removeOperation struct {
	Op   string `json:"op"`
	Path string `json:"path"`
}

func replace(path string, value interface{}) operation {
	return replaceOperation{
		Op:    "replace",
		Path:  path,
		Value: value,
	}
}

func remove(path string) operation {
	return removeOperation{
		Op:   "remove",
		Path: path,
	}
}

func patchApi(client kyma.Clientset, api kymaApi.Api, operations ...operation) (*kymaApi.Api, error) {

	if operations == nil {
		return nil, fmt.Errorf("no operations to perform in patch specified")
	}

	payloadBytes, _ := json.Marshal(operations)
	return client.GatewayV1alpha2().Apis(api.GetNamespace()).Patch(api.GetName(), types.JSONPatchType, payloadBytes)
}
