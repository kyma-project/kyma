package serverless

import (
	"fmt"
	"strings"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/apierror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/serverless/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockery -name=gqlFunctionConverter -output=automock -outpkg=automock -case=underscore
//go:generate failery -name=gqlFunctionConverter -case=underscore -output disabled -outpkg disabled
type gqlFunctionConverter interface {
	ToGQL(item *v1alpha1.Function) (*gqlschema.Function, error)
	ToGQLs(items []*v1alpha1.Function) ([]gqlschema.Function, error)
	ToFunction(name, namespace string, in gqlschema.FunctionMutationInput) (*v1alpha1.Function, error)
}

type functionConverter struct {
	extractor *functionUnstructuredExtractor
}

func newFunctionConverter() *functionConverter {
	return &functionConverter{
		extractor: newFunctionUnstructuredExtractor(),
	}
}

func (c *functionConverter) ToGQL(function *v1alpha1.Function) (*gqlschema.Function, error) {
	if function == nil {
		return nil, nil
	}

	envVariables := c.toGQLEnv(function.Spec.Env)
	resources := c.toGQLResources(function.Spec.Resources)
	replicas := c.toGQLReplicas(function.Spec.MinReplicas, function.Spec.MaxReplicas)
	status := c.getStatus(function.Status)

	return &gqlschema.Function{
		Name:         function.Name,
		Namespace:    function.Namespace,
		UID:          string(function.UID),
		Labels:       function.Labels,
		Source:       function.Spec.Source,
		Dependencies: function.Spec.Deps,
		Env:          envVariables,
		Replicas:     replicas,
		Resources:    resources,
		Status:       status,
	}, nil
}

func (c *functionConverter) ToGQLs(functions []*v1alpha1.Function) ([]gqlschema.Function, error) {
	if functions == nil {
		return nil, nil
	}

	var result []gqlschema.Function
	for _, function := range functions {
		converted, err := c.ToGQL(function)
		if err != nil {
			return nil, err
		}

		if converted != nil {
			result = append(result, *converted)
		}
	}
	return result, nil
}

func (c *functionConverter) ToFunction(name, namespace string, in gqlschema.FunctionMutationInput) (*v1alpha1.Function, error) {
	resources, errs := c.fromGQLResources(in.Resources)
	if len(errs) > 0 {
		err := apierror.NewInvalid(pretty.Function, errs)
		return nil, errors.Wrapf(err, "while converting to graphql resources field for %s [name: %s]. Resources: %v", pretty.Function, name, resources)
	}
	envVariables := c.fromGQLEnv(in.Env)
	minReplicas, maxReplicas := c.fromGQLReplicas(in.Replicas)

	return &v1alpha1.Function{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "serverless.kyma-project.io/v1alpha1",
			Kind:       "Function",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    in.Labels,
		},
		Spec: v1alpha1.FunctionSpec{
			Source:      in.Source,
			Deps:        in.Dependencies,
			Env:         envVariables,
			Resources:   resources,
			MinReplicas: minReplicas,
			MaxReplicas: maxReplicas,
		},
	}, nil
}

func (c *functionConverter) toGQLEnv(env []v1.EnvVar) []gqlschema.FunctionEnv {
	var variables []gqlschema.FunctionEnv
	for _, variable := range env {
		if variable.ValueFrom == nil {
			variables = append(variables, gqlschema.FunctionEnv{
				Name:  variable.Name,
				Value: variable.Value,
			})
			continue
		}

		configMapKeyRef := variable.ValueFrom.ConfigMapKeyRef
		secretKeyRef := variable.ValueFrom.SecretKeyRef
		if configMapKeyRef == nil && secretKeyRef == nil {
			continue
		}

		if configMapKeyRef != nil {
			variables = append(variables, gqlschema.FunctionEnv{
				Name: variable.Name,
				ValueFrom: &gqlschema.FunctionEnvValueFrom{
					Type:     gqlschema.FunctionEnvValueFromTypeConfigMap,
					Name:     configMapKeyRef.Name,
					Key:      configMapKeyRef.Key,
					Optional: configMapKeyRef.Optional,
				},
			})
			continue
		}
		variables = append(variables, gqlschema.FunctionEnv{
			Name: variable.Name,
			ValueFrom: &gqlschema.FunctionEnvValueFrom{
				Type:     gqlschema.FunctionEnvValueFromTypeSecret,
				Name:     secretKeyRef.Name,
				Key:      secretKeyRef.Key,
				Optional: secretKeyRef.Optional,
			},
		})
	}
	return variables
}

func (c *functionConverter) toGQLReplicas(minReplicas, maxReplicas *int32) gqlschema.FunctionReplicas {
	intPtr := func(ptrValue *int32) *int {
		var ptr *int
		if ptrValue != nil {
			value := int(*ptrValue)
			ptr = &value
		}
		return ptr
	}

	return gqlschema.FunctionReplicas{
		Min: intPtr(minReplicas),
		Max: intPtr(maxReplicas),
	}
}

func (c *functionConverter) toGQLResources(resources v1.ResourceRequirements) gqlschema.FunctionResources {
	stringPtr := func(str string) *string {
		return &str
	}
	extractResourceValues := func(item v1.ResourceList) gqlschema.ResourceValues {
		rv := gqlschema.ResourceValues{}
		if item, ok := item[v1.ResourceMemory]; ok {
			rv.Memory = stringPtr(item.String())
		}
		if item, ok := item[v1.ResourceCPU]; ok {
			rv.CPU = stringPtr(item.String())
		}
		return rv
	}

	gqlResources := gqlschema.FunctionResources{}
	if resources.Requests != nil {
		gqlResources.Requests = extractResourceValues(resources.Requests)
	}
	if resources.Limits != nil {
		gqlResources.Limits = extractResourceValues(resources.Limits)
	}
	return gqlResources
}

func (c *functionConverter) fromGQLEnv(env []gqlschema.FunctionEnvInput) []v1.EnvVar {
	var variables []v1.EnvVar
	for _, variable := range env {
		if variable.ValueFrom == nil {
			variables = append(variables, v1.EnvVar{
				Name:  variable.Name,
				Value: variable.Value,
			})
			continue
		}

		valueFrom := variable.ValueFrom
		if valueFrom.Type == gqlschema.FunctionEnvValueFromTypeConfigMap {
			variables = append(variables, v1.EnvVar{
				Name: variable.Name,
				ValueFrom: &v1.EnvVarSource{
					ConfigMapKeyRef: &v1.ConfigMapKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: valueFrom.Name,
						},
						Key:      valueFrom.Key,
						Optional: valueFrom.Optional,
					},
				},
			})
			continue
		}
		variables = append(variables, v1.EnvVar{
			Name: variable.Name,
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: valueFrom.Name,
					},
					Key:      valueFrom.Key,
					Optional: valueFrom.Optional,
				},
			},
		})
	}
	return variables
}

func (c *functionConverter) fromGQLReplicas(replicas gqlschema.FunctionReplicasInput) (*int32, *int32) {
	intPtr := func(ptrValue *int) *int32 {
		var ptr *int32
		if ptrValue != nil {
			value := int32(*ptrValue)
			ptr = &value
		}
		return ptr
	}

	return intPtr(replicas.Min), intPtr(replicas.Max)
}

func (c *functionConverter) fromGQLResources(resources gqlschema.FunctionResourcesInput) (v1.ResourceRequirements, apierror.ErrorFieldAggregate) {
	createResourceList := func(values gqlschema.ResourceValuesInput, pathPrefix string) (v1.ResourceList, apierror.ErrorFieldAggregate) {
		var resourcesList v1.ResourceList
		var errs apierror.ErrorFieldAggregate

		strMemory := ""
		if values.Memory != nil {
			strMemory = *values.Memory
		}
		memoryParsed, err := resource.ParseQuantity(strMemory)
		if strMemory != "" {
			if err != nil {
				errs = append(errs, apierror.NewInvalidField(fmt.Sprintf("%s.memory", pathPrefix), *values.Memory, "while parsing memory"))
			} else {
				resourcesList = v1.ResourceList{}
				resourcesList[v1.ResourceMemory] = memoryParsed
			}
		}

		strCPU := ""
		if values.CPU != nil {
			strCPU = *values.CPU
		}
		cpuParsed, err := resource.ParseQuantity(strCPU)
		if strCPU != "" {
			if err != nil {
				errs = append(errs, apierror.NewInvalidField(fmt.Sprintf("%s.cpu", pathPrefix), *values.CPU, "while parsing cpu"))
			} else {
				if resourcesList == nil {
					resourcesList = v1.ResourceList{}
				}
				resourcesList[v1.ResourceCPU] = cpuParsed
			}
		}

		return resourcesList, errs
	}

	resourcesReq := v1.ResourceRequirements{}
	var errs apierror.ErrorFieldAggregate

	requests, requestsErrs := createResourceList(resources.Requests, "resources.requests")
	resourcesReq.Requests = requests
	errs = append(errs, requestsErrs...)

	limits, limitsErrs := createResourceList(resources.Limits, "resources.limits")
	resourcesReq.Limits = limits
	errs = append(errs, limitsErrs...)

	return resourcesReq, errs
}

func (c *functionConverter) getStatus(status v1alpha1.FunctionStatus) gqlschema.FunctionStatus {
	conditions := status.Conditions

	// Initializing phase
	if len(conditions) == 0 {
		return gqlschema.FunctionStatus{
			Phase: gqlschema.FunctionPhaseTypeInitializing,
		}
	}

	functionConfigCreated := c.hasTrueType(v1alpha1.ConditionConfigurationReady, conditions)
	functionJobFinished := c.hasTrueType(v1alpha1.ConditionBuildReady, conditions)
	functionIsRunning := c.hasTrueType(v1alpha1.ConditionRunning, conditions)

	// Failed phase
	hasFailed, condition := c.getFailedCondition(conditions)
	if hasFailed {
		reasonType := c.getReasonType(condition.Type)
		if functionIsRunning {
			return gqlschema.FunctionStatus{
				Phase:   gqlschema.FunctionPhaseTypeNewRevisionError,
				Reason:  &reasonType,
				Message: &condition.Message,
			}
		}

		return gqlschema.FunctionStatus{
			Phase:   gqlschema.FunctionPhaseTypeFailed,
			Reason:  &reasonType,
			Message: &condition.Message,
		}
	}

	if functionConfigCreated {
		if functionJobFinished {
			// Running phase
			if functionIsRunning {
				return gqlschema.FunctionStatus{
					Phase: gqlschema.FunctionPhaseTypeRunning,
				}
			}

			// Deploying phase
			return gqlschema.FunctionStatus{
				Phase: gqlschema.FunctionPhaseTypeDeploying,
			}
		}

		// Building phase
		return gqlschema.FunctionStatus{
			Phase: gqlschema.FunctionPhaseTypeBuilding,
		}
	}

	// Otherwise Initializing phase
	return gqlschema.FunctionStatus{
		Phase: gqlschema.FunctionPhaseTypeInitializing,
	}
}

func (c *functionConverter) hasTrueType(conditionType v1alpha1.ConditionType, conditions []v1alpha1.Condition) bool {
	for _, cond := range conditions {
		if cond.Type == conditionType && cond.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}

func (c *functionConverter) getFailedCondition(conditions []v1alpha1.Condition) (bool, v1alpha1.Condition) {
	for _, cond := range conditions {
		if cond.Status == v1.ConditionFalse {
			return true, cond
		}
	}
	return false, v1alpha1.Condition{}
}

func (c *functionConverter) getReasonType(conditionType v1alpha1.ConditionType) gqlschema.FunctionReasonType {
	switch conditionType {
	case v1alpha1.ConditionConfigurationReady:
		return gqlschema.FunctionReasonTypeConfig
	case v1alpha1.ConditionBuildReady:
		return gqlschema.FunctionReasonTypeJob
	case v1alpha1.ConditionRunning:
		return gqlschema.FunctionReasonTypeService
	default:
		return gqlschema.FunctionReasonTypeConfig
	}
}

func (c *functionConverter) containsReason(reason v1alpha1.ConditionReason, subStrings []string) bool {
	reasonStr := string(reason)
	for _, subString := range subStrings {
		if strings.Contains(reasonStr, subString) {
			return true
		}
	}
	return false
}
