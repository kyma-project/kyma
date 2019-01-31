package testkit

import (
	kubelessV1 "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	kubeless "github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	acV1 "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	ac "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	subscriptionV1 "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
	subscription "github.com/kyma-project/kyma/components/event-bus/generated/push/clientset/versioned"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

type LambdaClient interface {
	DeployLambda(appName string) error
	CreateMapping(appName string) (*acV1.ApplicationMapping, error)
	CreateEventActivation(appName string) (*acV1.EventActivation, error)
	CreateSubscription(appName string) (*subscriptionV1.Subscription, error)
}

type lambdaClient struct {
	kubelessClientSet     *kubeless.Clientset
	subscriptionClientSet *subscription.Clientset
	appConnClientSet      *ac.Clientset
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

	appConnClientSet, err := ac.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &lambdaClient{
		kubelessClientSet:     kubelessClientSet,
		subscriptionClientSet: subscriptionClientSet,
		appConnClientSet:      appConnClientSet,
		namespace:             namespace,
	}, nil
}

func (c *lambdaClient) CreateMapping(appName string) (*acV1.ApplicationMapping, error) {
	am := &acV1.ApplicationMapping{
		TypeMeta:   metav1.TypeMeta{Kind: "ApplicationMapping", APIVersion: acV1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{Name: appName, Namespace: c.namespace},
	}

	return c.appConnClientSet.ApplicationconnectorV1alpha1().ApplicationMappings(c.namespace).Create(am)
}

func (c *lambdaClient) CreateEventActivation(appName string) (*acV1.EventActivation, error) {
	eaSpec := acV1.EventActivationSpec{
		DisplayName: "Commerce-events",
		SourceID:    appName,
	}

	ea := &acV1.EventActivation{
		TypeMeta:   metav1.TypeMeta{Kind: "EventActivation", APIVersion: acV1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{Name: appName, Namespace: c.namespace},

		Spec: eaSpec,
	}

	return c.appConnClientSet.ApplicationconnectorV1alpha1().EventActivations(c.namespace).Create(ea)
}

func (c *lambdaClient) CreateSubscription(appName string) (*subscriptionV1.Subscription, error) {
	subSpec := subscriptionV1.SubscriptionSpec{
		Endpoint:                      "http://e2e-lambda.production:8080",
		IncludeSubscriptionNameHeader: true,
		MaxInflight:                   400,
		PushRequestTimeoutMS:          2000,
		EventType:                     "order.created",
		EventTypeVersion:              "v1",
		SourceID:                      appName,
	}

	sub := &subscriptionV1.Subscription{
		TypeMeta:   metav1.TypeMeta{Kind: "Subscription", APIVersion: subscriptionV1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{Name: appName, Namespace: c.namespace, Labels: map[string]string{"Function": appName}},

		SubscriptionSpec: subSpec,
	}

	return c.subscriptionClientSet.EventingV1alpha1().Subscriptions(c.namespace).Create(sub)
}

func (c *lambdaClient) DeployLambda(appName string) error {
	lambda := c.createLambda(appName)

	_, err := c.kubelessClientSet.KubelessV1beta1().Functions(c.namespace).Create(lambda)
	if err != nil {
		return err
	}

	return nil
}

func (c *lambdaClient) createLambda(name string) *kubelessV1.Function {
	lambdaSpec := kubelessV1.FunctionSpec{
		Handler:             "e2e.entrypoint",
		Function:            "module.exports={'entrypoint':(event,context)=>{console.log(777)}}",
		FunctionContentType: "text",
		Runtime:             "nodejs8",
		Deps:                "",
		Deployment: v1beta1.Deployment{
			TypeMeta:   metav1.TypeMeta{Kind: "Deployment", APIVersion: v1beta1.SchemeGroupVersion.String()},
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: c.namespace, Labels: map[string]string{"function": name}},
			Spec: v1beta1.DeploymentSpec{
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: name,
							},
						},
					},
				},
			},
		},
		ServiceSpec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "http-function-port",
					Port:       8080,
					Protocol:   "TCP",
					TargetPort: intstr.FromInt(8080),
				},
			},
			Selector: map[string]string{"created-by": "kubeless", "function": name},
		},
		//HorizontalPodAutoscaler: v2beta1.HorizontalPodAutoscaler{},
	}

	return &kubelessV1.Function{
		TypeMeta:   metav1.TypeMeta{Kind: "Function", APIVersion: kubelessV1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: c.namespace},
		Spec:       lambdaSpec,
	}
}
