package runtimes

import (
	"fmt"
	"net/url"

	functionv1alpha1 "github.com/kyma-project/kyma/tests/function-controller/pkg/function/v1alpha1"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
)

type FunctionSimpleTestConfig struct {
	FnName       string
	Fn           *functionv1alpha1.Function
	InClusterURL *url.URL
}

func NewFunctionSimpleConfig(fnName string, toolBox shared.Container) (FunctionSimpleTestConfig, error) {
	inClusterURL, err := url.Parse(fmt.Sprintf("http://%s.%s.svc.cluster.local", fnName, toolBox.Namespace))
	if err != nil {
		return FunctionSimpleTestConfig{}, err
	}

	return FunctionSimpleTestConfig{
		FnName:       fnName,
		Fn:           functionv1alpha1.NewFunction(fnName, toolBox),
		InClusterURL: inClusterURL,
	}, nil
}
