package runtimes

import (
	"fmt"
	"math/rand"
	"net/url"

	"k8s.io/client-go/kubernetes"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/apirule"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/function"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/job"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/subscription"
)

type FunctionTestConfig struct {
	FnName       string
	APIRuleURL   *url.URL
	Fn           *function.Function
	ApiRule      *apirule.APIRule
	UsageName    string
	Subscription *subscription.Subscription
	Job          *job.Job
	InClusterURL *url.URL
	SinkURL      *url.URL
}

func NewFunctionConfig(fnName, usageKindName, domainName string, toolBox shared.Container, clientset *kubernetes.Clientset) (FunctionTestConfig, error) {
	gatewayURL, err := url.Parse(fmt.Sprintf("https://%s-%d.%s", fnName, rand.Uint32(), domainName))
	if err != nil {
		return FunctionTestConfig{}, err
	}

	inClusterURL, err := url.Parse(fmt.Sprintf("http://%s.%s.svc.cluster.local", fnName, toolBox.Namespace))
	if err != nil {
		return FunctionTestConfig{}, err
	}

	sinkURL, err := url.Parse(fmt.Sprintf("http://%s.%s.svc.cluster.local", fnName, toolBox.Namespace))
	if err != nil {
		return FunctionTestConfig{}, err
	}

	return FunctionTestConfig{
		FnName:       fnName,
		Fn:           function.NewFunction(fnName, toolBox),
		InClusterURL: inClusterURL,
		ApiRule:      apirule.New(fmt.Sprintf("%s-rule", fnName), toolBox),
		APIRuleURL:   gatewayURL,
		Subscription: subscription.New(fmt.Sprintf("%s-subscription", fnName), toolBox),
		Job:          job.New(fnName, clientset.BatchV1(), toolBox),
		SinkURL:      sinkURL,
	}, nil
}

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
