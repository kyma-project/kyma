package webhook

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	runtimeUtil "github.com/kyma-project/kyma/components/function-controller/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type FunctionCreateHandler struct {
	client  client.Client
	decoder *admission.Decoder
}

var (
	log = logf.Log.WithName("webhook")
)

const webhookEndpoint = "mutating-create-function"

// +kubebuilder:webhook:path=/mutating-create-function,mutating=true,failurePolicy=fail,groups=serverless.kyma-project.io,resources=functions,verbs=create;update,versions=v1alpha1,name=mfunction.kb.io

var _ inject.Client = &FunctionCreateHandler{}

// InjectClient injects the client into the FunctionCreateHandler
func (h *FunctionCreateHandler) InjectClient(c client.Client) error {
	h.client = c
	return nil
}

// InjectDecoder injects the decoder into the FunctionCreateHandler
func (h *FunctionCreateHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
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
	if obj.Spec.Visibility == "" {
		obj.Spec.Visibility = serverlessv1alpha1.FunctionVisibilityClusterLocal
	}


}

// Validate function values and return an error if the function is not valid
func (h *FunctionCreateHandler) validateFunction(obj *serverlessv1alpha1.Function, rnInfo *runtimeUtil.RuntimeInfo) field.ErrorList {
	errs := field.ErrorList{}

	errs = append(errs, h.validateFunctionMeta(&obj.ObjectMeta, field.NewPath("metadata"))...)
	errs = append(errs, h.validateFunctionSpec(&obj.Spec, rnInfo, field.NewPath("spec"))...)

	return errs
}

func (h *FunctionCreateHandler) validateFunctionMeta(meta *metav1.ObjectMeta, fldPath *field.Path) field.ErrorList {
	return apimachineryvalidation.ValidateObjectMeta(meta, true, apimachineryvalidation.NameIsDNS1035Label, fldPath)
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
		errs = append(errs, field.NotSupported(fldPath.Child("size"), spec.Size, functionSizes))
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
		errs = append(errs, field.NotSupported(fldPath.Child("runtime"), spec.Runtime, runtimes))
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
		errs = append(errs, field.NotSupported(fldPath.Child("functionContentType"), spec.FunctionContentType, functionContentTypes))
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
	value := os.Getenv(envName)
	if value == "" {
		return defaultValue
	}
	return value
}

// getRuntimeConfig returns the Function Controller ConfigMap from the cluster.
func (h *FunctionCreateHandler) getRuntimeConfig() (*corev1.ConfigMap, error) {
	cm := &corev1.ConfigMap{}

	err := h.client.Get(context.TODO(),
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
func (h *FunctionCreateHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	log.Info("received admission request", "request", req)

	fnConfig, err := h.getRuntimeConfig()
	if err != nil {
		log.Error(err, "Error reading controller configuration", "namespace", fnConfigNamespace, "name", fnConfigName)
		return admission.Errored(http.StatusInternalServerError, err)
	}

	rnInfo, err := runtimeUtil.New(fnConfig)
	if err != nil {
		log.Error(err, "Error creating RuntimeInfo", "namespace", fnConfig.Namespace, "name", fnConfig.Name)
		return admission.Errored(http.StatusBadRequest, err)
	}
	obj := &serverlessv1alpha1.Function{}

	if err := h.decoder.Decode(req, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	copyObj := obj.DeepCopy()

	// mutate values
	h.mutatingFunction(copyObj, rnInfo)

	// validate function and return an error describing the validation error if validation fails
	if errs := h.validateFunction(copyObj, rnInfo); len(errs) != 0 {
		return admission.Errored(http.StatusBadRequest, errs.ToAggregate())
	}

	marshalledFunction, err := json.Marshal(copyObj)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshalledFunction)
}
