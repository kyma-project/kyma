package runtimes

import (
	"fmt"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
)

func BasicPythonFunction(msg string, runtime serverlessv1alpha2.Runtime) serverlessv1alpha2.FunctionSpec {
	src := fmt.Sprintf(`import arrow 
def main(event, context):
	return "%s"`, msg)

	dpd := `requests==2.24.0
arrow==0.15.8`

	return serverlessv1alpha2.FunctionSpec{
		Runtime: runtime,
		Source: serverlessv1alpha2.Source{
			Inline: &serverlessv1alpha2.InlineSource{
				Source:       src,
				Dependencies: dpd,
			},
		},
		ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
			Function: &serverlessv1alpha2.ResourceRequirements{
				Profile: "M",
			},
			Build: &serverlessv1alpha2.ResourceRequirements{
				Profile: "normal",
			},
		},
	}
}

func BasicTracingPythonFunction(runtime serverlessv1alpha2.Runtime, externalURL string) serverlessv1alpha2.FunctionSpec {

	dpd := `opentelemetry-instrumentation==0.37b0
opentelemetry-instrumentation-requests==0.37b0
requests>=2.31.0`

	src := fmt.Sprintf(`import json

import requests
from opentelemetry.instrumentation.requests import RequestsInstrumentor


def main(event, context):
    RequestsInstrumentor().instrument()
    response = requests.get('%s', timeout=1)
    headers = response.request.headers
    tracingHeaders = {}
    for key, value in headers.items():
        if key.startswith("x-b3") or key.startswith("traceparent"):
            tracingHeaders[key] = value
    txtHeaders = json.dumps(tracingHeaders)
    return txtHeaders`, externalURL)

	return serverlessv1alpha2.FunctionSpec{
		Runtime: runtime,
		Source: serverlessv1alpha2.Source{
			Inline: &serverlessv1alpha2.InlineSource{
				Source:       src,
				Dependencies: dpd,
			},
		},
	}
}

func BasicPythonFunctionWithCustomDependency(msg string, runtime serverlessv1alpha2.Runtime) serverlessv1alpha2.FunctionSpec {
	src := fmt.Sprintf(
		`import arrow
def main(event, context):
	return "%s"`, msg)

	dpd := `requests==2.24.0
arrow==0.15.8
kyma-pypi-test==1.0.0`

	return serverlessv1alpha2.FunctionSpec{
		Runtime: runtime,
		Source: serverlessv1alpha2.Source{
			Inline: &serverlessv1alpha2.InlineSource{
				Source:       src,
				Dependencies: dpd,
			},
		},
		ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
			Function: &serverlessv1alpha2.ResourceRequirements{
				Profile: "M",
			},
			Build: &serverlessv1alpha2.ResourceRequirements{
				Profile: "normal",
			},
		},
	}
}

func BasicCloudEventPythonFunction(runtime serverlessv1alpha2.Runtime) serverlessv1alpha2.FunctionSpec {

	dpd := `requests>=2.31.0`

	src := `import json

def main(event, context):
    txtEventData = json.dumps(event.data)
    return txtEventData`

	return serverlessv1alpha2.FunctionSpec{
		Runtime: runtime,
		Source: serverlessv1alpha2.Source{
			Inline: &serverlessv1alpha2.InlineSource{
				Source:       src,
				Dependencies: dpd,
			},
		},
	}
}
