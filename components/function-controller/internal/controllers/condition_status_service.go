package controllers

import (
	corev1 "k8s.io/api/core/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func getSvcConditionStatus(svc *servingv1.Service) ConditionStatus {
	clen := len(svc.Status.Conditions)
	if clen < 1 {
		return ConditionStatusUnknown
	}

	cType := svc.Status.Conditions[clen-1].Type
	cStatus := svc.Status.Conditions[clen-1].Status

	if cStatus == corev1.ConditionFalse {
		return ConditionStatusFailed
	}

	if cType == "RoutesReady" && cStatus == corev1.ConditionTrue {
		return ConditionStatusSucceeded
	}

	return ConditionStatusUnknown
}
