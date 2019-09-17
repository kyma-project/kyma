/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mutating

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	runtimeUtil "github.com/kyma-project/kyma/components/function-controller/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	k8sTypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
)

var (
	log = logf.Log.WithName("webhook")
)

func init() {
	log.Info("init")
	webhookName := "mutating-create-function"
	if HandlerMap[webhookName] == nil {
		HandlerMap[webhookName] = []admission.Handler{}
	}
	HandlerMap[webhookName] = append(HandlerMap[webhookName], &FunctionCreateHandler{})
}

// FunctionCreateHandler handles Function
type FunctionCreateHandler struct {
	Client client.Client

	// Decoder decodes objects
	Decoder types.Decoder
}

// Mutates function values
func (h *FunctionCreateHandler) mutatingFunctionFn(obj *serverlessv1alpha1.Function, rnInfo *runtimeUtil.RuntimeInfo) {

	if obj.Spec.Size == "" {
		obj.Spec.Size = rnInfo.Defaults.Size
	}
	if obj.Spec.Runtime == "" {
		obj.Spec.Runtime = rnInfo.Defaults.Runtime
	}
	if obj.Spec.Timeout == 0 {
		obj.Spec.Timeout = rnInfo.Defaults.TimeOut
	}
	if obj.Spec.FunctionContentType == "" {
		obj.Spec.FunctionContentType = rnInfo.Defaults.FuncContentType
	}
}

// Validate function values and return an error if the function is not valid
func (h *FunctionCreateHandler) validateFunctionFn(obj *serverlessv1alpha1.Function, rnInfo *runtimeUtil.RuntimeInfo) error {

	// function size
	isValidFunctionSize := false
	var functionSizes []string
	for _, functionSize := range rnInfo.FuncSizes {
		functionSizes = append(functionSizes, functionSize.Size)

		if obj.Spec.Size == functionSize.Size {
			isValidFunctionSize = true
			break
		}
	}

	if !isValidFunctionSize {
		return fmt.Errorf("Passed function size: '%v', but size should be one of '%v'", obj.Spec.Size, strings.Join(functionSizes, ", "))
	}

	// function serverless
	isValidRuntime := false
	var runtimes []string
	for _, runtime := range rnInfo.AvailableRuntimes {
		runtimes = append(runtimes, runtime.ID)
		if obj.Spec.Runtime == runtime.ID {
			isValidRuntime = true
			break
		}
	}

	if !isValidRuntime {
		return fmt.Errorf("Passed Runtime: '%v', but runtime should be one of '%v'", obj.Spec.Runtime, strings.Join(runtimes, ", "))
	}

	// function content type
	isValidFunctionContentType := false
	var functionContentTypes []string
	for _, functionContentType := range rnInfo.FuncTypes {
		functionContentTypes = append(functionContentTypes, functionContentType.Type)
		if obj.Spec.FunctionContentType == functionContentType.Type {
			isValidFunctionContentType = true
			break
		}
	}
	if !isValidFunctionContentType {
		return fmt.Errorf("Passed functionContetype: '%v', but functionContentType should be one of '%v'", obj.Spec.FunctionContentType, strings.Join(functionContentTypes, ", "))
	}

	return nil
}

var _ admission.Handler = &FunctionCreateHandler{}

var (
	// name of function config
	fnConfigName = getEnvDefault("CONTROLLER_CONFIGMAP", "fn-config")

	// namespace of function config
	fnConfigNamespace = getEnvDefault("CONTROLLER_CONFIGMAP_NS", "default")
)

func getEnvDefault(envName string, defaultValue string) string {
	// use default value if environment variable is empty
	var value string
	if value = os.Getenv(envName); value == "" {
		return defaultValue
	}
	return value
}
func (h *FunctionCreateHandler) getRuntimeConfig(fnConfig *corev1.ConfigMap) error {
	// Get Function Controller Configuration
	err := h.Client.Get(context.TODO(), k8sTypes.NamespacedName{Name: fnConfigName, Namespace: fnConfigNamespace}, fnConfig)
	if err != nil {
		log.Error(err, "Unable to read Function controller's configuration", "namespace", fnConfigNamespace, "name", fnConfigName)
		return err
	}
	return nil
}

// Handle handles admission requests.
func (h *FunctionCreateHandler) Handle(ctx context.Context, req types.Request) types.Response {
	log.Info("received admission request", "request", req)

	fnConfig := &corev1.ConfigMap{}
	if err := h.getRuntimeConfig(fnConfig); err != nil {
		return admission.ErrorResponse(http.StatusInternalServerError, err)
	}
	rnInfo, err := runtimeUtil.New(fnConfig)
	if err != nil {
		log.Error(err, "Error while trying to get a new RuntimeInfo instance", "namespace", fnConfig.Namespace, "name", fnConfig.Name)
		return admission.ErrorResponse(http.StatusBadRequest, err)
	}
	obj := &serverlessv1alpha1.Function{}

	err = h.Decoder.Decode(req, obj)
	if err != nil {
		return admission.ErrorResponse(http.StatusBadRequest, err)
	}
	copy := obj.DeepCopy()

	// mutate values
	h.mutatingFunctionFn(copy, rnInfo)

	// validate function and return an error describing the validation error if validation fails
	err = h.validateFunctionFn(copy, rnInfo)
	if err != nil {
		return admission.ErrorResponse(http.StatusBadRequest, err)
	}

	return admission.PatchResponse(obj, copy)
}

var _ inject.Client = &FunctionCreateHandler{}

// InjectClient injects the client into the FunctionCreateHandler
func (h *FunctionCreateHandler) InjectClient(c client.Client) error {
	h.Client = c
	return nil
}

var _ inject.Decoder = &FunctionCreateHandler{}

// InjectDecoder injects the decoder into the FunctionCreateHandler
func (h *FunctionCreateHandler) InjectDecoder(d types.Decoder) error {
	h.Decoder = d
	return nil
}
