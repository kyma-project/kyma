package testsuite

import (
	"fmt"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"

	serverless "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/retry"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
)

const functionFmt = `	
const expectedPayload = "%s";	
const legacy = %t;	
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
function sendReq(event, url, resolve, reject) {
	const options = {
		body: event,
		uri: url,
		json: true,
		headers: {
			'content-type': 'application/cloudevents+json'
		}
	}
	console.log(JSON.stringify(options));
	request.post(options, (error, response, body) => {	
		if (error) {	
			reject(error);	
		}	
		resolve(response) ;	
	});
}
function getGateway() {
	if (legacy) {
		return process.env.GATEWAY_URL;
	} else {
		let envKey = Object.keys(process.env).find(val => val.endsWith('_GATEWAY_URL'));
		return process.env[envKey]
	}
}
function prepareEvent(event){
	return {
		"specversion": event.extensions.request.headers["ce-specversion"],
		"source": event.extensions.request.headers["ce-source"],
		"type": event.extensions.request.headers["ce-type"],
		"eventtypeversion": event.extensions.request.headers["ce-eventtypeversion"],
		"id": event.extensions.request.headers["ce-id"],
		"data" : event.data 
	}
}

module.exports = { main: function (event, context) {	
	console.log("==============================");	
	console.log("Legacy:           ", legacy)
	console.log("Expected Payload: ", expectedPayload)
	console.log("==============================");
	console.log("Received event: ");	
	console.log(JSON.stringify(event, null, 2));	
	console.log("==============================");	
	console.log("Received context:");	
	console.log(JSON.stringify(context, null, 2));	
	console.log("==============================");	
	
	const gateway = getGateway()
    if (gateway === undefined) {
		throw new Error("gateway is undefined")	
	}
    if (event["data"] !== expectedPayload) {	
		throw new Error("Payload not as expected")	
    }
	return new Promise((resolve, reject) => {	
		const url = gateway + "/ce";
		var preparedEvent = prepareEvent(event);
		console.log("Counter URL: ", url);	
		sendReq(preparedEvent, url, resolve, reject);	
	}).then(resolved, rejected);	
} };	
`
const functionDeps = `{"dependencies":{"request": "^2.88.0", "circular-json": "^0.5.9"}}`

// DeployFunction deploys function to the cluster. The function will do PUT /counter to connected application upon receiving
// an event
type DeployFunction struct {
	client          dynamic.ResourceInterface
	name            string
	port            int
	expectedPayload string
	legacy          bool
}

var _ step.Step = &DeployFunction{}

func NewDeployFunction(name, expectedPayload string, port int, client dynamic.ResourceInterface, legacy bool) *DeployFunction {
	return &DeployFunction{
		client:          client,
		name:            name,
		port:            port,
		legacy:          legacy,
		expectedPayload: expectedPayload,
	}
}

func (s *DeployFunction) Name() string {
	return fmt.Sprintf("Deploy function %s", s.name)
}

func (s *DeployFunction) Run() error {
	if err := s.createFunction(); err != nil {
		return err
	}

	if err := retry.Do(s.isFunctionReady); err != nil {
		return errors.Wrap(err, "function not ready")
	}

	return nil
}

func (s *DeployFunction) createFunction() error {
	function := &serverless.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: serverless.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: s.name,
		},
		Spec: serverless.FunctionSpec{
			Source: fmt.Sprintf(functionFmt, s.expectedPayload, s.legacy),
			Deps:   functionDeps,
		},
	}

	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(function)
	if err != nil {
		return err
	}

	_, err = s.client.Create(&unstructured.Unstructured{Object: unstructuredObj}, metav1.CreateOptions{})
	return err
}

func (s *DeployFunction) Cleanup() error {
	err := s.client.Delete(s.name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return helpers.AwaitResourceDeleted(func() (interface{}, error) {
		return s.client.Get(s.name, metav1.GetOptions{})
	})
}

func (s *DeployFunction) isFunctionReady() error {
	function, err := s.getFunction()
	if err != nil {
		return err
	}

	if !s.isReady(function) {
		return fmt.Errorf("function is not ready yet: %+v \n", function)
	}

	return nil
}

func (s *DeployFunction) getFunction() (serverless.Function, error) {
	functionUnstructed, err := s.client.Get(s.name, metav1.GetOptions{})
	if err != nil {
		return serverless.Function{}, err
	}

	if len(functionUnstructed.Object) == 0 {
		return serverless.Function{}, errors.New("no function found")
	}

	var function serverless.Function
	runtime.DefaultUnstructuredConverter.FromUnstructured(functionUnstructed.Object, &function)
	if err != nil {
		return serverless.Function{}, err
	}

	return function, nil
}

func (s *DeployFunction) isReady(fn serverless.Function) bool {
	conditions := fn.Status.Conditions
	if len(conditions) == 0 {
		return false
	}

	for _, condition := range conditions {
		if condition.Type == serverless.ConditionRunning {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}
