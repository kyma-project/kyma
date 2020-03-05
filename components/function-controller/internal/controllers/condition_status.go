package controllers

import (
	serverless "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	knapis "knative.dev/pkg/apis"
)

//go:generate stringer -type=ConditionStatus -trimprefix=ConditionStatus

type ConditionStatus int8

const (
	ConditionStatusSucceeded ConditionStatus = iota + 1
	ConditionStatusFailed
	ConditionStatusUnknown
)

// function controller should rebuild image if function last action
// resulted with error
func conditionCheck(cts []serverless.Condition) rebuildImg {
	ctslen := len(cts)

	if ctslen < 1 {
		return rebuildImgRequired
	}
	return cts[ctslen-1].Type == serverless.ConditionTypeError
}

func getConditionStatus(con []knapis.Condition) ConditionStatus {
	cLen := len(con)

	if cLen < 1 {
		return ConditionStatusUnknown
	}

	cType := con[0].Type
	cStatus := con[0].Status

	if cStatus == corev1.ConditionFalse {
		return ConditionStatusFailed
	}

	if cType == knapis.ConditionSucceeded && cStatus == corev1.ConditionTrue {
		return ConditionStatusSucceeded
	}

	return ConditionStatusUnknown
}
