package testkit

import (
	kubelessV1 "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	kubeless "github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/rest"
)

type LambdaClient interface {
	DeployLambda(appName string) error
	DeleteLambda(appName string, options *metav1.DeleteOptions) error
}

type lambdaClient struct {
	kubelessClientSet *kubeless.Clientset
	namespace         string
}

func NewLambdaClient(config *rest.Config, namespace string) (LambdaClient, error) {
	kubelessClientSet, err := kubeless.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &lambdaClient{
		kubelessClientSet: kubelessClientSet,
		namespace:         namespace,
	}, nil
}

func (c *lambdaClient) DeployLambda(appName string) error {
	lambda := c.createLambda(appName)

	_, err := c.kubelessClientSet.KubelessV1beta1().Functions(c.namespace).Create(lambda)
	if err != nil {
		return err
	}

	return nil
}

func (c *lambdaClient) DeleteLambda(appName string, options *metav1.DeleteOptions) error {
	return c.kubelessClientSet.KubelessV1beta1().Functions(c.namespace).Delete(appName, options)
}

func (c *lambdaClient) createLambda(name string) *kubelessV1.Function {
	//TODO: Modify lambda's code - it can handle URL to gist
	lambdaSpec := kubelessV1.FunctionSpec{
		Handler:             "e2e.entrypoint",
		Function:            "module.exports={'entrypoint':(event,context)=>{console.log(\"IMPLEMENT_ME\")}}",
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
	}

	return &kubelessV1.Function{
		TypeMeta:   metav1.TypeMeta{Kind: "Function", APIVersion: kubelessV1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: c.namespace},
		Spec:       lambdaSpec,
	}
}
