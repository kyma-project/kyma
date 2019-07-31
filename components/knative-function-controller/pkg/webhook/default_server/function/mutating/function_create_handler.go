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
	"strings"

	runtimev1alpha1 "github.com/kyma-project/kyma/components/knative-function-controller/pkg/apis/runtime/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
)

var (
	functionSizes        = []string{"S", "M", "L", "XL"}
	runtimes             = []string{"nodejs6", "nodejs8"}
	functionContentTypes = []string{"plaintext", "base64"}
	log                  = logf.Log.WithName("webhook")
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
func (h *FunctionCreateHandler) mutatingFunctionFn(obj *runtimev1alpha1.Function) {
	if obj.Spec.Size == "" {
		obj.Spec.Size = "S"
	}
	if obj.Spec.Runtime == "" {
		obj.Spec.Runtime = "nodejs8"
	}
	if obj.Spec.Timeout == 0 {
		obj.Spec.Timeout = 180
	}
	if obj.Spec.FunctionContentType == "" {
		obj.Spec.FunctionContentType = "plaintext"
	}
}

// Validate function values and return an error if the function is not valid
func (h *FunctionCreateHandler) validateFunctionFn(obj *runtimev1alpha1.Function) error {
	// function size
	isValidFunctionSize := false
	for _, functionSize := range functionSizes {
		if obj.Spec.Size == functionSize {
			isValidFunctionSize = true
			break
		}
	}
	if !isValidFunctionSize {
		return fmt.Errorf("size should be one of '%v'", strings.Join(functionSizes, ","))
	}

	// function runtime
	isValidRuntime := false
	for _, runtime := range runtimes {
		if obj.Spec.Runtime == runtime {
			isValidRuntime = true
			break
		}
	}
	if !isValidRuntime {
		return fmt.Errorf("runtime should be one of '%v'", strings.Join(runtimes, ","))
	}

	// function content type
	isValidFunctionContentType := false
	for _, functionContentType := range functionContentTypes {
		if obj.Spec.FunctionContentType == functionContentType {
			isValidFunctionContentType = true
			break
		}
	}
	if !isValidFunctionContentType {
		return fmt.Errorf("functionContentType should be one of '%v'", strings.Join(functionContentTypes, ","))
	}

	return nil
}

var _ admission.Handler = &FunctionCreateHandler{}

// Handle handles admission requests.
func (h *FunctionCreateHandler) Handle(ctx context.Context, req types.Request) types.Response {
	log.Info("received admission request", "request", req)

	obj := &runtimev1alpha1.Function{}

	err := h.Decoder.Decode(req, obj)
	if err != nil {
		return admission.ErrorResponse(http.StatusBadRequest, err)
	}
	copy := obj.DeepCopy()

	// mutate values
	h.mutatingFunctionFn(copy)

	// validate function and return an error describing the validation error if validation fails
	err = h.validateFunctionFn(copy)
	if err != nil {
		return admission.ErrorResponse(http.StatusInternalServerError, err)
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
