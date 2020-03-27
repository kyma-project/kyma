package testsuite

import (
	"fmt"

	kubelessv1beta1 "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	kubelessclientset "github.com/kubeless/kubeless/pkg/client/clientset/versioned/typed/kubeless/v1beta1"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/retry"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
)

const lambdaFunctionFmt = `
const expectedPayload = "%s";
const request = require('request');
const JSON = require('circular-json');

function resolved(result) {
  console.log('Resolved:');
  console.log(JSON.stringify(result, null, 2));
}
	
function rejected(result) {
  console.log('Rejected:');
  console.log(JSON.stringify(result, null, 2));
}

function sendReq(url, resolve, reject) {
    request.post(url, { json: true }, (error, response, body) => {
        if (error) {
            reject(error);
        }
        resolve(response) ;
    });
}

module.exports = { main: function (event, context) {
	console.log("Received event: ");
	console.log(JSON.stringify(event, null, 2));
	console.log("==============================");
	console.log("Received context:");
	console.log(JSON.stringify(context, null, 2));
	console.log("==============================");
	
	const gatewayUrlEnvKey =  Object.keys(process.env).find(val => val.endsWith('_GATEWAY_URL'))
	if (gatewayUrlEnvKey === undefined) {
		throw new Error("Environmental variable with '_GATEWAY_URL' suffix is undefined")
	}
	console.log("Gateway URL Env Key:", gatewayUrlEnvKey);
	
	const gatewayUrl = process.env[gatewayUrlEnvKey]
	console.log("Gateway URL:", gatewayUrl);
    if (gatewayUrl === undefined) {
		throw new Error("Environmental variable with Gateway URL is empty");
	}
    
    if (event["data"] !== expectedPayload) {
		throw new Error("Payload not as expected");
    }
	
	return new Promise((resolve, reject) => {
		const url = gatewayUrl + "/counter";
		console.log("Counter URL: ", url);
		sendReq(url, resolve, reject);
	}).then(resolved, rejected);
} };
`

const legacyLambdaFunctionFmt = `
const expectedPayload = "%s";
const request = require('request');
const JSON = require('circular-json');

function resolved(result) {
  console.log('Resolved:');
  console.log(JSON.stringify(result, null, 2));
}
	
function rejected(result) {
  console.log('Rejected:');
  console.log(JSON.stringify(result, null, 2));
}

function sendReq(url, resolve, reject) {
    request.post(url, { json: true }, (error, response, body) => {
        if (error) {
            reject(error);
        }
        resolve(response) ;
    });
}

module.exports = { main: function (event, context) {
	console.log("Received event: ");
	console.log(JSON.stringify(event, null, 2));
	console.log("==============================");
	console.log("Received context:");
	console.log(JSON.stringify(context, null, 2));
	console.log("==============================");
    if (process.env.GATEWAY_URL === undefined) {
		throw new Error("GATEWAY_URL is undefined")
	}
    
    if (event["data"] !== expectedPayload) {
		throw new Error("Payload not as expected")
    }
	
	return new Promise((resolve, reject) => {
		const url = process.env.GATEWAY_URL + "/counter";
		console.log("Counter URL: ", url);
		sendReq(url, resolve, reject);
	}).then(resolved, rejected);
} };
`

// DeployLambda deploys lambda to the cluster. The lambda will do PUT /counter to connected application upon receiving
// an event
type DeployLambda struct {
	*helpers.LambdaHelper
	functions          kubelessclientset.FunctionInterface
	name               string
	port               int
	expectedPayload    string
	lambdaFunctionCode string
}

var _ step.Step = &DeployLambda{}

// NewDeployLambda returns new DeployLambda
func NewDeployLambda(name, expectedPayload string, port int, functions kubelessclientset.FunctionInterface, pods coreclient.PodInterface, legacy bool) *DeployLambda {
	return &DeployLambda{
		LambdaHelper:       helpers.NewLambdaHelper(pods),
		functions:          functions,
		name:               name,
		port:               port,
		expectedPayload:    expectedPayload,
		lambdaFunctionCode: lambdaFunctionCode(legacy),
	}
}

// Name returns name name of the step
func (s *DeployLambda) Name() string {
	return fmt.Sprintf("Deploy lambda %s", s.name)
}

// Run executes the step
func (s *DeployLambda) Run() error {
	lambda := s.createLambda()
	_, err := s.functions.Create(lambda)
	if err != nil {
		return err
	}

	err = retry.Do(s.isLambdaReady)
	if err != nil {
		return errors.Wrap(err, "lambda function not ready")
	}

	return nil
}

func (s *DeployLambda) createLambda() *kubelessv1beta1.Function {
	lambdaSpec := kubelessv1beta1.FunctionSpec{
		Handler:             "handler.main",
		Function:            fmt.Sprintf(s.lambdaFunctionCode, s.expectedPayload),
		FunctionContentType: "text",
		Runtime:             "nodejs8",
		Deps:                `{"dependencies":{"request": "^2.88.0", "circular-json": "^0.5.9"}}`,
		Deployment: appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:   s.name,
				Labels: map[string]string{"function": s.name},
			},
			Spec: appsv1.DeploymentSpec{
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: s.name,
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
					TargetPort: intstr.FromInt(s.port),
				},
			},
			Selector: map[string]string{"created-by": "kubeless", "function": s.name},
		},
	}

	return &kubelessv1beta1.Function{
		ObjectMeta: metav1.ObjectMeta{Name: s.name},
		Spec:       lambdaSpec,
	}
}

// Cleanup removes all resources that may possibly created by the step
func (s *DeployLambda) Cleanup() error {
	err := s.functions.Delete(s.name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return retry.Do(
		s.isLambdaTerminated)
}

func (s *DeployLambda) isLambdaReady() error {
	pods, err := s.ListLambdaPods(s.name)
	if err != nil {
		return err
	}

	if len(pods) == 0 {
		return errors.New("no function pods found")
	}

	for _, pod := range pods {
		if !helpers.IsPodReady(pod) {
			return errors.New("pod is not ready yet")
		}
	}

	return nil
}

func (s *DeployLambda) isLambdaTerminated() error {
	pods, err := s.ListLambdaPods(s.name)
	if err != nil {
		return err
	}

	if len(pods) != 0 {
		return errors.New("function pods found")
	}

	return nil
}

func lambdaFunctionCode(legacy bool) string {
	if legacy {
		return legacyLambdaFunctionFmt
	}

	return lambdaFunctionFmt
}
