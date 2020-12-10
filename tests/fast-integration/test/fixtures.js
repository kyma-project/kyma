const commerceMockYaml = `
apiVersion: v1
kind: Namespace
metadata:
  labels:
    istio-injection: enabled
  name: mocks
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: commerce-mock
  namespace: mocks
  labels:
    app: commerce-mock
spec:
  selector:
    matchLabels:
      app: commerce-mock
  strategy:
    rollingUpdate:
      maxUnavailable: 1
  replicas: 1
  template:
    metadata:
      labels:
        app: commerce-mock
    spec:
      containers:
#      - image: eu.gcr.io/kyma-project/xf-application-mocks/commerce-mock:latest
      - image: pbochynski/commerce-mock-lite:0.3
        imagePullPolicy: Always
        name: commerce-mock
        ports:
        - name: http
          containerPort: 10000
        env:
        - name: DEBUG
          value: "true"
        - name: RENEWCERT_JOB_CRON
          value: "00 00 */12 * * *"
        # volumeMounts:
        # - mountPath: "/app/keys"
        #   name: commerce-mock-volume
        resources:
          requests:
            memory: "150Mi"
            cpu: "10m"
          limits:
            memory: "350Mi"
            cpu: "300m"
      # volumes:
      # - name: commerce-mock-volume
      #   persistentVolumeClaim:
      #     claimName: commerce-mock 
---
apiVersion: v1
kind: Service
metadata:
  name: commerce-mock
  namespace: mocks
  labels:
    app: commerce-mock
spec:
  ports:
  - name: http
    port: 10000
  selector:
    app: commerce-mock
---
apiVersion: gateway.kyma-project.io/v1alpha1
kind: APIRule
metadata:
  name: commerce-mock
  namespace: mocks
spec:
  gateway: kyma-gateway.kyma-system.svc.cluster.local
  rules:
  - accessStrategies:
    - config: {}
      handler: allow
    methods: ["*"]
    path: /.*
  service:
    host: commerce
    name: commerce-mock
    port: 10000`;

const appConnectorYaml = `
apiVersion: applicationconnector.kyma-project.io/v1alpha1
kind: Application
metadata:
  name: commerce
spec:
  description: Commerce mock
---
apiVersion: applicationconnector.kyma-project.io/v1alpha1
kind: ApplicationMapping
metadata:
  name: commerce
---
apiVersion: serverless.kyma-project.io/v1alpha1
kind: Function
metadata:
  labels:
    serverless.kyma-project.io/build-resources-preset: local-dev
    serverless.kyma-project.io/function-resources-preset: S
    serverless.kyma-project.io/replicas-preset: S
  name: lastorder
spec:
  deps: '{ "name": "orders", "version": "1.0.0", "dependencies": {"axios": "^0.19.2"}}'
  maxReplicas: 1
  minReplicas: 1
  source: |
    let lastOrder = {};
    const axios = require('axios');
    async function getOrder(code) {
      let url = process.env.GATEWAY_URL+"/site/orders/"+code;
      console.log("URL: %s", url); 
      let response = await axios.get(url,{headers:{"X-B3-Sampled":1}})
      console.log(response.data); 
      return response.data;}
    module.exports = { 
      main: function (event, context) {
        if (event.data && event.data.orderCode){
          lastOrder = getOrder(event.data.orderCode)
        }
        return lastOrder;
     }}
---
apiVersion: gateway.kyma-project.io/v1alpha1
kind: APIRule
metadata:
  name: lastorder
spec:
  gateway: kyma-gateway.kyma-system.svc.cluster.local
  rules:
  - accessStrategies:
    - config: {}
      handler: allow
    methods: ["*"]
    path: /.*
  service:
    host: lastorder
    name: lastorder
    port: 80
---
apiVersion: eventing.knative.dev/v1alpha1
kind: Trigger
metadata:
  labels:
    function: lastorder
  name: function-lastorder
spec:
  broker: default
  filter:
    attributes:
      eventtypeversion: v1
      source: commerce
      type: order.created
  subscriber:
    ref:
      apiVersion: v1
      kind: Service
      name: lastorder`;

const tokenRequestYaml = `
apiVersion: applicationconnector.kyma-project.io/v1alpha1
kind: TokenRequest
metadata:
  name: commerce`;

const genericServiceClass = (name, namespace) => `
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceClass
metadata:
  name: ${name}
  namespace: ${namespace}
`;

const serviceCatalogResources = (
  webServicesExternalName,
  eventsExternalName
) => `
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: commerce-webservices
spec:
  serviceClassExternalName: ${webServicesExternalName}
---
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: commerce-events
spec:
  serviceClassExternalName: ${eventsExternalName}
---
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceBinding
metadata:
  labels:
    function: lastorder
  name: commerce-lastorder-binding
spec:
  instanceRef:
    name: commerce-webservices
---
apiVersion: servicecatalog.kyma-project.io/v1alpha1
kind: ServiceBindingUsage
metadata:
  labels:
    function: lastorder
    serviceBinding: commerce-lastorder-binding
  name: commerce-lastorder-sbu
spec:
  serviceBindingRef:
    name: commerce-lastorder-binding
  usedBy:
    kind: serverless-function
    name: lastorder`;

module.exports = {
  commerceMockYaml,
  appConnectorYaml,
  tokenRequestYaml,
  genericServiceClass,
  serviceCatalogResources,
};
