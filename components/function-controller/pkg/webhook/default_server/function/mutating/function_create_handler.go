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
	"net/http"
	"os"

	corev1 "k8s.io/api/core/v1"
	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	apivalidation "k8s.io/kubernetes/pkg/apis/core/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	runtimeUtil "github.com/kyma-project/kyma/components/function-controller/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
func (h *FunctionCreateHandler) mutatingFunction(obj *serverlessv1alpha1.Function, rnInfo *runtimeUtil.RuntimeInfo) {
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
func (h *FunctionCreateHandler) validateFunction(obj *serverlessv1alpha1.Function, rnInfo *runtimeUtil.RuntimeInfo) field.ErrorList {
	errs := field.ErrorList{}

	errs = append(errs, h.validateFunctionMeta(obj.ObjectMeta, rnInfo, field.NewPath("metadata"))...)
	errs = append(errs, h.validateFunctionSpec(obj.Spec, rnInfo, field.NewPath("spec"))...)

	return errs
}

func (h *FunctionCreateHandler) validateFunctionMeta(meta *metav1.ObjectMeta, fldPath *field.Path) field.ErrorList {
	return apivalidation.ValidateObjectMeta(meta, true, apimachineryvalidation.NameIsDNS1035Label, fldPath)
}

func (h *FunctionCreateHandler) validateFunctionSpec(spec *serverlessv1alpha1.FunctionSpec, rnInfo *runtimeUtil.RuntimeInfo, fldPath *field.Path) field.ErrorList {
	errs := field.ErrorList{}

	// function size
	isValidFunctionSize := false
	var functionSizes []string
	for _, functionSize := range rnInfo.FuncSizes {
		functionSizes = append(functionSizes, functionSize.Size)
		if spec.Size == functionSize.Size {
			isValidFunctionSize = true
			break
		}
	}

	if !isValidFunctionSize {
		return append(errs, field.NotSupported(fldPath.Child("size"), spec.Size, functionSizes))
	}

	// function serverless
	isValidRuntime := false
	var runtimes []string
	for _, runtime := range rnInfo.AvailableRuntimes {
		runtimes = append(runtimes, runtime.ID)
		if spec.Runtime == runtime.ID {
			isValidRuntime = true
			break
		}
	}

	if !isValidRuntime {
		return append(errs, field.NotSupported(fldPath.Child("runtime"), spec.Runtime, runtimes))
	}

	// function content type
	isValidFunctionContentType := false
	var functionContentTypes []string
	for _, functionContentType := range rnInfo.FuncTypes {
		functionContentTypes = append(functionContentTypes, functionContentType.Type)
		if spec.FunctionContentType == functionContentType.Type {
			isValidFunctionContentType = true
			break
		}
	}

	if !isValidFunctionContentType {
		return append(errs, field.NotSupported(fldPath.Child("functionContentType"), spec.FunctionContentType, functionContentTypes))
	}

	return errs
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

// getRuntimeConfig returns the Function Controller ConfigMap from the cluster.
func (h *FunctionCreateHandler) getRuntimeConfig() (*corev1.ConfigMap, error) {
	cm := &corev1.ConfigMap{}

	err := h.Client.Get(context.TODO(),
		client.ObjectKey{
			Name:      fnConfigName,
			Namespace: fnConfigNamespace,
		},
		cm,
	)

	if err != nil {
		return nil, err
	}
	return cm, nil
}

// Handle handles admission requests.
func (h *FunctionCreateHandler) Handle(ctx context.Context, req types.Request) types.Response {
	log.Info("received admission request", "request", req)

	fnConfig, err := h.getRuntimeConfig()
	if err != nil {
		log.Error(err, "Error reading controller configuration", "namespace", fnConfigNamespace, "name", fnConfigName)
		return admission.ErrorResponse(http.StatusInternalServerError, err)
	}

	rnInfo, err := runtimeUtil.New(fnConfig)
	if err != nil {
		log.Error(err, "Error creating RuntimeInfo", "namespace", fnConfig.Namespace, "name", fnConfig.Name)
		return admission.ErrorResponse(http.StatusBadRequest, err)
	}
	obj := &serverlessv1alpha1.Function{}

	if err := h.Decoder.Decode(req, obj); err != nil {
		return admission.ErrorResponse(http.StatusBadRequest, err)
	}
	copy := obj.DeepCopy()

	// mutate values
	h.mutatingFunction(copy, rnInfo)

	// validate function and return an error describing the validation error if validation fails
	if errs := h.validateFunction(copy, rnInfo); len(errs) != 0 {
		return admission.ErrorResponse(http.StatusBadRequest, errs.ToAggregate())
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
