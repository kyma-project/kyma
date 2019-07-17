package testsuite

import (
	"github.com/avast/retry-go"
	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	kubelessClient "github.com/kubeless/kubeless/pkg/client/clientset/versioned/typed/kubeless/v1beta1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/consts"
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

type DeployLambda struct {
	*helpers.LambdaHelper
	functions kubelessClient.FunctionInterface
}

var _ step.Step = &DeployLambda{}

func NewDeployLambda(functions kubelessClient.FunctionInterface, pods coreClient.PodInterface) *DeployLambda {
	return &DeployLambda{
		LambdaHelper: helpers.NewLambdaHelper(pods),
		functions:    functions,
	}
}

func (s *DeployLambda) Name() string {
	return "Deploy lambda"
}

func (s *DeployLambda) Run() error {
	lambda := s.createLambda()

	_, err := s.functions.Create(lambda)
	if err != nil {
		return err
	}

	err = retry.Do(s.isLambdaReady, retry.Delay(200*time.Millisecond))
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
				Name:   consts.AppName,
				Labels: map[string]string{"function": consts.AppName},
			},
			Spec: extensionsApi.DeploymentSpec{
				Template: coreApi.PodTemplateSpec{
					Spec: coreApi.PodSpec{
						Containers: []coreApi.Container{
							{
								Name: consts.AppName,
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
					TargetPort: intstr.FromInt(8080),
				},
			},
			Selector: map[string]string{"created-by": "kubeless", "function": consts.AppName},
		},
	}

	return &kubelessApi.Function{
		ObjectMeta: metav1.ObjectMeta{Name: consts.AppName},
		Spec:       lambdaSpec,
	}
}

// Cleanup removes all resources that may possibly created by the step
func (s *DeployLambda) Cleanup() error {
	err := s.functions.Delete(consts.AppName, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return retry.Do(s.isLambdaTerminated, retry.Delay(200*time.Millisecond))
}

func (s *DeployLambda) isLambdaReady() error {
	pods, err := s.ListLambdaPods()
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
	pods, err := s.ListLambdaPods()
	if err != nil {
		return err
	}

	if len(pods) != 0 {
		return errors.New("function pods found")
	}

	return nil
}
