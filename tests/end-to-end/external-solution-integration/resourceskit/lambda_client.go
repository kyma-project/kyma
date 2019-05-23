package resourceskit

import (
	kubelessV1 "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	kubeless "github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/consts"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/rest"
)

const lambdaFunction = `
	const request = require('request')
	module.exports = { 'entrypoint': function (event, context) {
		return new Promise((resolve, reject) => {
			const url = "http://" + process.env.GATEWAY_URL + "/counter"
			sendReq(url, resolve, reject)
		})
	} }
	function sendReq(url, resolve, reject) {
        request.post(url, { json: true }, (error, response, body) => {
            if (error) {
                resolve(error)
            }
            resolve(response) 
        })
    }`

type LambdaClient interface {
	DeployLambda() error
	DeleteLambda() error
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

func (c *lambdaClient) DeployLambda() error {
	lambda := c.createLambda()

	_, err := c.kubelessClientSet.KubelessV1beta1().Functions(c.namespace).Create(lambda)
	if err != nil {
		return err
	}

	return nil
}

func (c *lambdaClient) DeleteLambda() error {
	return c.kubelessClientSet.KubelessV1beta1().Functions(c.namespace).Delete(consts.AppName, &metav1.DeleteOptions{})
}

func (c *lambdaClient) createLambda() *kubelessV1.Function {
	lambdaSpec := kubelessV1.FunctionSpec{
		Handler:             "e2e.entrypoint",
		Function:            lambdaFunction,
		FunctionContentType: "text",
		Runtime:             "nodejs8",
		Deps:                "request",
		Deployment: v1beta1.Deployment{
			TypeMeta:   metav1.TypeMeta{Kind: "Deployment", APIVersion: v1beta1.SchemeGroupVersion.String()},
			ObjectMeta: metav1.ObjectMeta{Name: consts.AppName, Namespace: c.namespace, Labels: map[string]string{"function": consts.AppName}},
			Spec: v1beta1.DeploymentSpec{
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: consts.AppName,
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
			Selector: map[string]string{"created-by": "kubeless", "function": consts.AppName},
		},
	}

	return &kubelessV1.Function{
		TypeMeta:   metav1.TypeMeta{Kind: "Function", APIVersion: kubelessV1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{Name: consts.AppName, Namespace: c.namespace},
		Spec:       lambdaSpec,
	}
}
