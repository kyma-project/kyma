# Function Fails to Start With Customized OpenTelemetry Dependencies

## Symptom

After customizing OpenTelemetry dependencies, the Serverless Function is not starting.

## Cause

Users can configure their own dependencies as part of the Function's dependencies. While this allows for customization, it also introduces the risk of dependency conflicts, especially with OpenTelemetry packages, as those have strict version requirements for their dependencies. For example, `opentelemetry-instrumentation` requires specific versions of `opentelemetry-api` and `opentelemetry-sdk`. A mismatch in versions can lead to runtime errors.
Downgrading or upgrading one package without aligning the others can lead to runtime errors or import issues.


## Solution

Avoid downgrading OpenTelemetry packages in your Function's configuration. Use the dependencies provided by the Serverless module for [Python](https://raw.githubusercontent.com/kyma-project/serverless/refs/heads/main/components/runtimes/python/python312/requirements.txt) and [Node.js](https://raw.githubusercontent.com/kyma-project/serverless/refs/heads/main/components/runtimes/nodejs/nodejs22/package.json). These dependencies are maintained and upgraded regularly. They are also tested to ensure compatibility. If you need to customize OpenTelemetry dependencies, ensure that all related packages (e.g., `opentelemetry-api`, `opentelemetry-sdk`, `opentelemetry-instrumentation`) are aligned to compatible versions. Refer to the [PyPi](https://pypi.org/) and [npm](https://www.npmjs.com/) pages to verify compatibility of the desired version.