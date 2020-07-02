## Test external application

#### Create and label a namespace
```bash
kubectl create ns test
kubectl label ns test knative-eventing-injection=enabled --overwrite
```

#### Create an application
```bash
cat << EOF | kubectl apply -f -
apiVersion: applicationconnector.kyma-project.io/v1alpha1
kind: Application
metadata:
    name: "test-app"
spec:
    accessLabel: "test-app"
EOF
```

#### Create an application mapping
```bash
cat << EOF | kubectl apply -f -
apiVersion: applicationconnector.kyma-project.io/v1alpha1
kind: ApplicationMapping
metadata:
    name: "test-app"
    namespace: "test"
EOF
```

#### Register events API to the application connector
```bash
## Install brew install one-click-integration
one-click-integration.sh \
           -r "test-app" \
           -c "${KUBECONFIG}"
```

#### Register API in Kyma
> Wait until the pods in kyma-integration are ready
```bash
curl -v \
    --cert generated.pem -k \
    -H "Content-Type: application/json" \
    -d @<(curl https://raw.githubusercontent.com/kyma-incubator/examples/b22e12dacb02f2dddeedcab35cc6ab3acd47cc98/cloud-events-e2e-scenario/register-events.json) \
    --silent --show-error --fail \
    "https://gateway.${DOMAIN}/test-app/v1/metadata/services"
```

#### Create a serviceinstance
```bash
## Check the serviceinstance created
service_class_line=$(kubectl get serviceclasses.servicecatalog.k8s.io -A | grep es-all-events | tail -n +1 )
echo "selected: ${service_class_line}"
service_class=$(awk '{print $3}' <<< "$service_class_line")
echo "\nfound service class for eventing: $service_class"

cat << EOF | kubectl apply -f -
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  annotations:
    tags: ""
  name: test-app
  namespace: test
spec:
  serviceClassExternalName: ${service_class}
  servicePlanExternalName: default
EOF
```

#### Create a function
```bash
cat << EOF | kubectl apply -f -
apiVersion: serverless.kyma-project.io/v1alpha1
kind: Function
metadata:
  name: test
  namespace: test
spec:
  deps: "{ \n  \"name\": \"{${APP_NAME}\",\n  \"version\": \"1.0.0\",\n  \"dependencies\":
    {}\n}"
  source: |
    module.exports = { main: function (event, context) {
        console.log(\`event = \${JSON.stringify(event.data)}\`);
        console.log(\`headers = \${JSON.stringify(event.extensions.request.headers)}\`);
    } }
EOF
```

#### Create a trigger
```bash
cat << EOF | kubectl apply -f -
apiVersion: eventing.knative.dev/v1alpha1
kind: Trigger
metadata:
  name: test
  namespace: test
spec:
  broker: default
  filter:
    attributes:
      eventtypeversion: v1
      source: test-app
      type: stage.com.external.solution.order.created
  subscriber:
    ref:
      apiVersion: v1
      kind: Service
      name: test
EOF
```

#### Send an event to function
```bash
curl -k -v --cert generated.crt --key generated.key \
        -H "Content-Type: text" \
                        -X POST https://gateway.local.kyma.pro/test-app/v1/events \
        -d '{"source-id": "test-app", "event-type": "stage.com.external.solution.order.created", "event-type-version": "v1", "event-time": "2018-11-02T22:08:41+00:00", "data": { "test": "hello installation wg" }}'
```
