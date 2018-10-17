---
title: Calling registered service from within Kyma
type: Getting Started
---

This guide shows how to call the registered service from within the Kyma using a simple lambda function.


## Prerequisites

- Remote Environment created and bound to the `production` Environment.
- Client certificates for the RE generated.


## Steps

1. Register a service with the following specyfication to the desierd Remote Envoronment:
```json
{
  "name": "Ec without events",
  "provider": "hybris",
  "Identifier": "aa112bc",
  "description": "This is some EC!3",
  "api": {
    "targetUrl": "http://httpbin.org/",
    "spec": {
      "swagger":"2.0"
    }
  }
}
```
Our service will call http://httpbin.org.
Save the received service id, as it is used in the later steps.

2. Next you need to create the Service Instance. To achive this you need the `externalName` of the Cluster Service Class.
To get the `externalName` run:
```
kubectl get clusterserviceclass {SERVICE_ID}  -o jsonpath='{.spec.externalName}'
```

Use it to create the Service Instance
```
cat <<EOF | kubectl apply -f -
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: my-service-instance-name
  namespace: production
spec:
  clusterServiceClassExternalName: {EXTERNAL_NAME}
EOF
```

3. Create lambda that calls registered service
```
cat <<EOF | kubectl apply -f -
apiVersion: kubeless.io/v1beta1
kind: Function
metadata:
  name: my-lambda
  namespace: production
spec:
  deployment:
    spec:
      template:
        metadata:
          labels:
            re-{RE_NAME}-{SERVICE_ID}: "true"
        spec:
          containers:
          - name: ""
            resources: {}
  deps: |-
    {
        "name": "example-1",
        "version": "0.0.1",
        "dependencies": {
          "request": "^2.85.0"
        }
    }
  function: |-
    const request = require('request');

    module.exports = { main: function (event, context) {
        return new Promise((resolve, reject) => {
            const url = \`\${process.env.GATEWAY_URL}/uuid\`;
            const options = {
                url: url,
            };
              
            sendReq(url, resolve, reject)
        })
    } }

    function sendReq(url, resolve, reject) {
        request.get(url, { json: true }, (error, response, body) => {
            if(error){
                resolve(error);
            }
            console.log("Response acquired succesfully! Uuid: " + response.body.uuid);
            resolve(response);
        })
    }
  function-content-type: text
  handler: handler.main
  horizontalPodAutoscaler:
    spec:
      maxReplicas: 0
  runtime: nodejs8
  service:
    ports:
    - name: http-function-port
      port: 8080
      protocol: TCP
      targetPort: 8080
    selector:
      created-by: kubeless
      function: my-lambda
  timeout: ""
  topic: http
EOF
```
The lambda will call the service with additional path `/uuid`

4. Create Service Binding and Service Binding Usage to bind the previously create Service Instance to the lambda.

```
cat <<EOF | kubectl apply -f -
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceBinding
metadata:
  labels:
    Function: my-lambda
  name: my-service-binding
  namespace: production
spec:
  instanceRef:
    name: my-service-instance-name
EOF
```

```
cat <<EOF | kubectl apply -f -
apiVersion: servicecatalog.kyma.cx/v1alpha1
kind: ServiceBindingUsage
metadata:
  labels:
    Function: my-lambda
    ServiceBinding: my-service-binding
  name: my-service-binding
  namespace: production
spec:
  serviceBindingRef:
    name: my-service-binding
  usedBy:
    kind: function
    name: my-lambda
EOF
```

5. To expose lambda outside the cluster create a Virtual Service:
```
cat <<EOF | kubectl apply -f -
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  labels:
    apiName: my-lambda
    apiNamespace: production
  name: my-lambda
  namespace: production
spec:
  gateways:
  - kyma-gateway.kyma-system.svc.cluster.local
  hosts:
  - my-lambda-production.{CLUSTER_DOMAIN}
  http:
  - match:
    - uri:
        regex: /.*
    route:
    - destination:
        host: my-lambda.production.svc.cluster.local
        port:
          number: 8080
EOF
```

6. To verify that everything was setup correctly you now can call the lambda through https:
```
curl https://my-lambda.{CLUSTER_DOMAIN}/ -k
```

On Minikube one additional step is required.
For you to be able to access the lambda over https you need to edit `/etc/hosts/` file and map `https://my-lambda-production.kyma.local` to the Minikube ip.

After that you can call the lambda
```
curl https://my-lambda-production.kyma.local/ -k
```
