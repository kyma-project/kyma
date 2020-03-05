package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
)

const (
	FnUUID   = "fnUUID"
	ImageTag = "imageTag"
)

func (fn *Function) GetSanitizedDeps() string {
	if fn.Spec.Deps != "" {
		return fn.Spec.Deps
	}
	return "{}"
}

func (fn *Function) ImgLabelSelector() labels.Selector {
	return labels.SelectorFromSet(
		map[string]string{
			FnUUID:   string(fn.UID),
			ImageTag: fn.Status.ImageTag,
		},
	)
}

func (fn *Function) LabelSelector() labels.Selector {
	return labels.SelectorFromSet(
		map[string]string{
			FnUUID: string(fn.UID),
		},
	)
}

func (fn *Function) NamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      fn.Name,
		Namespace: fn.Namespace,
	}
}

func (fn *Function) ConditionReasonCreateConfigSucceeded(imageTag string) *FunctionStatus {
	condition := Condition{
		Type:               ConditionTypeInitialized,
		Reason:             ConditionReasonCreateConfigSucceeded,
		LastTransitionTime: metav1.Now(),
	}
	return &FunctionStatus{
		Phase:              FunctionPhaseBuilding,
		ImageTag:           imageTag,
		ObservedGeneration: fn.GetGeneration(),
		Conditions:         append(fn.Status.Conditions, condition),
	}
}

func (fn *Function) FunctionPhaseFailed(reason ConditionReason, msg string) *FunctionStatus {
	condition := Condition{
		Type:               ConditionTypeError,
		Reason:             reason,
		Message:            msg,
		LastTransitionTime: metav1.Now(),
	}
	return &FunctionStatus{
		Phase:              FunctionPhaseFailed,
		ObservedGeneration: fn.GetGeneration(),
		ImageTag:           fn.Status.ImageTag,
		Conditions:         append(fn.Status.Conditions, condition),
	}
}

func (fn *Function) ConditionReasonUnknown() *FunctionStatus {
	return fn.FunctionPhaseFailed(ConditionReasonUnknown, "")
}

func (fn *Function) ConditionReasonCreateConfigFailed(err error) *FunctionStatus {
	return fn.FunctionPhaseFailed(ConditionReasonCreateConfigFailed, err.Error())
}

func (fn *Function) ConditionReasonUpdateServiceSucceeded(msg string) *FunctionStatus {
	condition := Condition{
		Type:               ConditionTypeDeploying,
		Reason:             ConditionReasonUpdateServiceSucceeded,
		Message:            msg,
		LastTransitionTime: metav1.Now(),
	}
	return &FunctionStatus{
		Phase:              FunctionPhaseDeploying,
		ObservedGeneration: fn.GetGeneration(),
		ImageTag:           fn.Status.ImageTag,
		Conditions:         append(fn.Status.Conditions, condition),
	}
}

func (fn *Function) FunctionStatusDeploying(reason ConditionReason, msg string) *FunctionStatus {
	condition := Condition{
		Type:               ConditionTypeDeploying,
		Reason:             reason,
		Message:            msg,
		LastTransitionTime: metav1.Now(),
	}
	return &FunctionStatus{
		Phase:              FunctionPhaseDeploying,
		ObservedGeneration: fn.GetGeneration(),
		ImageTag:           fn.Status.ImageTag,
		Conditions:         append(fn.Status.Conditions, condition),
	}
}

func (fn *Function) FunctionStatusGetConfigFailed(err error) *FunctionStatus {
	return fn.FunctionPhaseFailed(ConditionReasonGetConfigFailed, err.Error())
}

func (fn *Function) FunctionStatusBuildSucceed() *FunctionStatus {
	condition := Condition{
		Type:               ConditionTypeImageCreated,
		Reason:             ConditionReasonBuildSucceeded,
		LastTransitionTime: metav1.Now(),
	}
	return &FunctionStatus{
		Phase:              FunctionPhaseDeploying,
		ObservedGeneration: fn.GetGeneration(),
		ImageTag:           fn.Status.ImageTag,
		Conditions:         append(fn.Status.Conditions, condition),
	}
}

func (fn *Function) FunctionStatusBuildRunning() *FunctionStatus {
	return &FunctionStatus{
		Phase:              FunctionPhaseBuilding,
		ObservedGeneration: fn.GetGeneration(),
		ImageTag:           fn.Status.ImageTag,
		Conditions:         fn.Status.Conditions,
	}
}

func (fn *Function) FunctionStatusUpdateRuntimeConfig() *FunctionStatus {
	condition := Condition{
		Type:               ConditionTypeInitialized,
		Reason:             ConditionReasonUpdateRuntimeConfig,
		LastTransitionTime: metav1.Now(),
	}
	return &FunctionStatus{
		Phase:              FunctionPhaseBuilding,
		ImageTag:           fn.Status.ImageTag,
		ObservedGeneration: fn.GetGeneration(),
		Conditions:         append(fn.Status.Conditions, condition),
	}
}

func (fn *Function) FunctionStatusUpdateConfigSucceeded(imageTag string) *FunctionStatus {
	condition := Condition{
		Type:               ConditionTypeInitialized,
		Reason:             ConditionReasonUpdateConfigSucceeded,
		LastTransitionTime: metav1.Now(),
	}
	return &FunctionStatus{
		Phase:              FunctionPhaseBuilding,
		ImageTag:           imageTag,
		ObservedGeneration: fn.GetGeneration(),
		Conditions:         append(fn.Status.Conditions, condition),
	}
}

func (fn *Function) FunctionStatusUpdateConfigFailed(err error) *FunctionStatus {
	return fn.FunctionPhaseFailed(ConditionReasonUpdateConfigFailed, err.Error())
}

func (fn *Function) FunctionStatusInitializing() *FunctionStatus {
	return &FunctionStatus{
		Phase:              FunctionPhaseInitializing,
		ObservedGeneration: fn.GetGeneration(),
		Conditions:         fn.Status.Conditions,
		ImageTag:           fn.Status.ImageTag,
	}
}

func (fn *Function) FunctionStatusDeploySucceeded() *FunctionStatus {
	condition := Condition{
		Type:               ConditionTypeDeployed,
		Reason:             ConditionReasonDeploySucceeded,
		LastTransitionTime: metav1.Now(),
	}
	return &FunctionStatus{
		Phase:              FunctionPhaseRunning,
		ImageTag:           fn.Status.ImageTag,
		ObservedGeneration: fn.GetGeneration(),
		Conditions:         append(fn.Status.Conditions, condition),
	}
}
