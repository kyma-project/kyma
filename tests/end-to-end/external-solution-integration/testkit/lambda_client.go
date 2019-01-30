package testkit

import (
	kubeless "github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	ea "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	subscription "github.com/kyma-project/kyma/components/event-bus/generated/push/clientset/versioned"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

type LambdaClient interface {
	CreateLambda()
}

type lambdaClient struct {
	kubelessClientSet     *kubeless.Clientset
	subscriptionClientSet *subscription.Clientset
	eaClientSet           *ea.Clientset
	namespace             string
}

func NewLambdaClient(namespace string) (LambdaClient, error) {
	kubeconfig := os.Getenv("KUBECONFIG")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	kubelessClientSet, err := kubeless.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	subscriptionClientSet, err := subscription.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	eaClientSet, err := ea.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &lambdaClient{
		kubelessClientSet:     kubelessClientSet,
		subscriptionClientSet: subscriptionClientSet,
		eaClientSet:           eaClientSet,
		namespace:             namespace,
	}, nil
}

func (c *lambdaClient) CreateLambda() {

}
