# Service Catalog and Service Broker Tests

## Overview

This folder contains the tests which check if a given service broker integrates with the Service Catalog. Each of these tests checks if the Cluster service classes contain all services exposed by the broker.

## Usage

This section explains how to add a test for another broker and how to run a Service Catalog test on a local machine.

To test your changes and build an image, use the `make build-image` command with **DOCKER_PUSH_REPOSITORY** and **DOCKER_PUSH_DIRECTORY** variables, for example:
```
DOCKER_PUSH_REPOSITORY=eu.gcr.io DOCKER_PUSH_DIRECTORY=/kyma-project/develop make build-image
```

### Adding a test for another broker

The test checks if the Service Catalog contains all the necessary service classes. All actions use standard APIs, such as Kubernetes and the Open Service Broker API, for all tests, and only the service broker URL changes.

### Running a Service Catalog test on a local machine

Follow these steps to run a Service Catalog test on a local machine:

1. Run Kyma.
2. Modify the broker service, using the `kubectl edit` command:
 - add `name: http` to the **port** section
 - add the `auth.istio.io/80: NONE` annotation

3. Create an Ingress to expose the service broker. For example:

 ```yaml
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  annotations:
    kubernetes.io/ingress.class: istio
  generation: 1
  name: helm-broker-ing
  namespace: kyma-system
spec:
  rules:
  - host: helm-broker.kyma
    http:
      paths:
      - backend:
          serviceName: helm-broker
          servicePort: 80
        path: /.*
   tls:
    - secretName: istio-ingress-certs
```

4. Update the `/etc/hosts` file with a given host value. For example:

  `127.0.0.1 helm-broker.kyma`
5. Create files that look like Secrets:

  `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt`

  You can copy the content of the file in any Pod. For example:
```bash
kc exec -it helm-broker-6cd7cc8697-c8wks -n kyma-system -- cat /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
```

6. Create a secret local file:
```
  kc describe serviceaccount service-catalog-controller-manager -n kyma-system
```

  It gives you an output which looks like this:
```
Name:         service-catalog-controller-manager
Namespace:    kyma-system
Labels:       <none>
Annotations:  <none>
Image pull secrets:  <none>
Mountable secrets:   service-catalog-controller-manager-token-jd98k
Tokens:              service-catalog-controller-manager-token-jd98k
Events:  <none>
```

  Get the token value:
```
kc describe secret service-catalog-controller-manager-token-jd98k -n kyma-system
```

  Rewrite the token value to the `/var/run/secrets/kubernetes.io/serviceaccount/token` file.
