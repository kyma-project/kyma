package controllers

import (
	serverless "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

//go:generate stringer -type=FunctionEventType -trimprefix=FunctionEventType

type FunctionEventType int8

const (
	FunctionEventTypeWarning FunctionEventType = iota + 1
	Normal
)

func (r *FunctionReconciler) recordPhaseChange(
	fn *serverless.Function,
	fnEvtType FunctionEventType,
	phase serverless.StatusPhase) {
	r.recorder.Event(fn, fnEvtType.String(), string(phase), "phase changed")
}
