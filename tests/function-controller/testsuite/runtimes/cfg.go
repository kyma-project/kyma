package runtimes

import (
	"fmt"
	"net/url"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/function"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
)

type FunctionSimpleTestConfig struct {
	FnName       string
	Fn           *function.Function
	InClusterURL *url.URL
}

func NewFunctionSimpleConfig(fnName string, toolBox shared.Container) (FunctionSimpleTestConfig, error) {
	inClusterURL, err := url.Parse(fmt.Sprintf("http://%s.%s.svc.cluster.local", fnName, toolBox.Namespace))
	if err != nil {
		return FunctionSimpleTestConfig{}, err
	}

	return FunctionSimpleTestConfig{
		FnName:       fnName,
		Fn:           function.NewFunction(fnName, toolBox),
		InClusterURL: inClusterURL,
	}, nil
}
