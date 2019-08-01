package testsuite

import (
	"fmt"
	"github.com/avast/retry-go"
	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	kubelessClient "github.com/kubeless/kubeless/pkg/client/clientset/versioned/typed/kubeless/v1beta1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/pkg/errors"
	coreApi "k8s.io/api/core/v1"
	extensionsApi "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	coreClient "k8s.io/client-go/kubernetes/typed/core/v1"
	"time"
)

const lambdaFunction = `
const request = require('request');

function resolved(result) {
  console.log('Resolved', result);
}
	
function rejected(result) {
  console.log("Rejected", result);
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
	console.log("Received event: ", event);
	
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
	functions kubelessClient.FunctionInterface
	name      string
	port      int
}

var _ step.Step = &DeployLambda{}

// NewDeployLambda returns new DeployLambda
func NewDeployLambda(name string, port int, functions kubelessClient.FunctionInterface, pods coreClient.PodInterface) *DeployLambda {
	return &DeployLambda{
		LambdaHelper: helpers.NewLambdaHelper(pods),
		functions:    functions,
		name:         name,
		port:         port,
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

	err = retry.Do(s.isLambdaReady, retry.Delay(500*time.Millisecond))
	if err != nil {
		return errors.Wrap(err, "lambda function not ready")
	}

	return nil
}

func (s *DeployLambda) createLambda() *kubelessApi.Function {
	lambdaSpec := kubelessApi.FunctionSpec{
		Handler:             "handler.main",
		Function:            lambdaFunction,
		FunctionContentType: "text",
		Runtime:             "nodejs8",
		Deps:                `{"dependencies":{"request": "^2.88.0"}}`,
		Deployment: extensionsApi.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:   s.name,
				Labels: map[string]string{"function": s.name},
			},
			Spec: extensionsApi.DeploymentSpec{
				Template: coreApi.PodTemplateSpec{
					Spec: coreApi.PodSpec{
						Containers: []coreApi.Container{
							{
								Name: s.name,
							},
						},
					},
				},
			},
		},
		ServiceSpec: coreApi.ServiceSpec{
			Ports: []coreApi.ServicePort{
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

	return &kubelessApi.Function{
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

	return retry.Do(s.isLambdaTerminated, retry.Delay(200*time.Millisecond))
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
