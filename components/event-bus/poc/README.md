## [POC] Run EventBus on top of Knative NatssChannel CRD

### Export GKE cluster properties
```bash
export CLUSTER_NAME="sayan-event-bus" \
export GCP_PROJECT="kyma-project" \
export GCP_ZONE="europe-west1-b" \
export CLUSTER_VERSION="1.12.7"
```

### Create a GKE cluster
```bash
gcloud container --project "$GCP_PROJECT" clusters create "$CLUSTER_NAME" \
--addons=HorizontalPodAutoscaling,HttpLoadBalancing \
--machine-type "n1-standard-4" \
--cluster-version "$CLUSTER_VERSION" --zone "$GCP_ZONE" \
--enable-stackdriver-kubernetes --enable-ip-alias \
--enable-autoscaling --min-nodes=1 --max-nodes=10 \
--enable-autorepair \
--scopes cloud-platform
```

### Grant cluster-admin permissions
```bash
kubectl create clusterrolebinding cluster-admin-binding \
  --clusterrole=cluster-admin \
  --user=$(gcloud config get-value core/account)
```

### Install Istio
```bash
kubectl apply --filename https://raw.githubusercontent.com/knative/serving/v0.8.0/third_party/istio-1.1.7/istio-crds.yaml

curl -L https://raw.githubusercontent.com/knative/serving/v0.8.0/third_party/istio-1.1.7/istio.yaml \
  | kubectl apply -f -

# label the default namespace with istio-injection=enabled.
kubectl label namespace default istio-injection=enabled

# watch pods until is shows STATUS of Running or Completed
kubectl get pods --namespace istio-system -w
```

### Install Knative
```bash
kubectl apply --selector knative.dev/crd-install=true \
   --filename https://github.com/knative/serving/releases/download/v0.8.0/serving.yaml \
   --filename https://github.com/knative/eventing/releases/download/v0.8.0/release.yaml

kubectl apply --filename https://github.com/knative/serving/releases/download/v0.8.0/serving.yaml --selector networking.knative.dev/certificate-provider!=cert-manager \
   --filename https://github.com/knative/eventing/releases/download/v0.8.0/release.yaml
```

### Install NATSS server
```bash
kubectl create namespace natss; \
kubectl apply -n natss -f https://raw.githubusercontent.com/knative/eventing/v0.8.0/contrib/natss/config/broker/natss.yaml
```

### Install NATSS controller and dispatcher
```bash
# export your docker registry and use ko to install NATSS controller and dispatcher
kubectl apply -n knative-eventing -f https://github.com/knative/eventing/releases/download/v0.8.0/natss.yaml
```

### Remove default configmap for CCP

```bash
kubectl delete cm default-channel-webhook -n knative-eventing
```

### Add a default configmap for Channel
```bash
cat << EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: default-ch-webhook
  namespace: knative-eventing
data:
  default-ch-config: |
    clusterDefault:
      apiVersion: messaging.knative.dev/v1alpha1
      kind: NatssChannel 
EOF
```

### Install EventBus CRD
```bash
# install required CRDs
kubectl apply \
   -f https://raw.githubusercontent.com/kyma-project/kyma/release-1.4/resources/cluster-essentials/templates/event-activation.crd.yaml \
   -f https://raw.githubusercontent.com/kyma-project/kyma/release-1.4/resources/cluster-essentials/templates/eventing-subscription.crd.yaml

```


### (Optional)Create new images for the EventBus apps
```bash
# export your docker registry to push images
export DOCKER_REGISTRY=shazra;
cd $GOPATH/src/github.com/kyma-project/kyma/components/event-bus/; \
make clean build-image; \
docker tag event-bus-publish-knative $DOCKER_REGISTRY/event-bus-publish-knative:latest; \
docker push $DOCKER_REGISTRY/event-bus-publish-knative; \
docker tag event-bus-subscription-controller-knative $DOCKER_REGISTRY/event-bus-subscription-controller-knative:latest; \
docker push $DOCKER_REGISTRY/event-bus-subscription-controller-knative 0
```

### Install event-bus
```bash
cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Namespace
metadata:
  labels:
    istio-injection: enabled
  name: kyma-system
---
# Source: event-bus/charts/publish-knative/templates/service-account.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: event-bus-publish
  namespace: kyma-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: event-bus-publish-knative
rules:
  - apiGroups:
      - eventing.knative.dev
    resources:
      - channels
    verbs:
      - get
      - list
  - apiGroups:
      - messaging.knative.dev
    resources:
      - channels
      - channels/status
    verbs:
      - get
      - list
      - watch
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: event-bus-publish-knative
  namespace: kyma-system
subjects:
  - kind: ServiceAccount
    name: event-bus-publish
    namespace: kyma-system
roleRef:
  kind: ClusterRole
  name: event-bus-publish-knative
  apiGroup: rbac.authorization.k8s.io
---
# Source: event-bus/charts/subscription-controller-knative/templates/service-account.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: event-bus-subscription-controller-knative-sa
  namespace: kyma-system
---
### Kyma subscriptions ########################################################
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: event-bus-kyma-subscriptions-controller
rules:
  - apiGroups:
      - eventing.kyma-project.io
    resources:
      - subscriptions
      - subscriptions/status
    verbs:
      - get
      - watch
      - list
      - update
  - apiGroups:
      - messaging.knative.dev
    resources:
      - channels
      - channels/status
    verbs: ['*']
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: event-bus-kyma-subscriptions-controller
subjects:
  - kind: ServiceAccount
    name: event-bus-subscription-controller-knative-sa
    namespace: kyma-system
roleRef:
  kind: ClusterRole
  name: event-bus-kyma-subscriptions-controller
  apiGroup: rbac.authorization.k8s.io
---
### Kyma event-activations ####################################################
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: event-bus-kyma-eventactivations-controller
rules:
  - apiGroups: ["applicationconnector.kyma-project.io"]
    resources: ["eventactivations"]
    verbs: ["get", "watch", "list", "update"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: event-bus-kyma-eventactivations-controller
subjects:
  - kind: ServiceAccount
    name: event-bus-subscription-controller-knative-sa
    namespace: kyma-system
roleRef:
  kind: ClusterRole
  name: event-bus-kyma-eventactivations-controller
  apiGroup: rbac.authorization.k8s.io
---
### Knative channels ##########################################################
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: event-bus-knative-channels-controller
rules:
  - apiGroups: ["eventing.knative.dev"]
    resources: ["channels"]
    verbs: ["get", "create", "update", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: event-bus-knative-channels-controller
  namespace: kyma-system
subjects:
  - kind: ServiceAccount
    name: event-bus-subscription-controller-knative-sa
    namespace: kyma-system
roleRef:
  kind: ClusterRole
  name: event-bus-knative-channels-controller
  apiGroup: rbac.authorization.k8s.io
---
### Knative subscriptions #####################################################
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: event-bus-knative-subscriptions-controller
rules:
  - apiGroups: ["eventing.knative.dev"]
    resources: ["subscriptions"]
    verbs: ["get", "create", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: event-bus-knative-subscriptions-controller
  namespace: kyma-system
subjects:
  - kind: ServiceAccount
    name: event-bus-subscription-controller-knative-sa
    namespace: kyma-system
roleRef:
  kind: ClusterRole
  name: event-bus-knative-subscriptions-controller
  apiGroup: rbac.authorization.k8s.io
---
### Events #####################################################
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: event-bus-events-controller
rules:
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["create", "update", "patch"]
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: event-bus-events-controller
  namespace: kyma-system
subjects:
  - kind: ServiceAccount
    name: event-bus-subscription-controller-knative-sa
    namespace: kyma-system
roleRef:
  kind: ClusterRole
  name: event-bus-events-controller
  apiGroup: rbac.authorization.k8s.io
---
# Source: event-bus/templates/tests/test-e2e-tester.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name:  test-event-bus-tester
  namespace: kyma-system
  labels:
    helm-chart-test: "true"
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: test-event-bus-subs
  labels:
    helm-chart-test: "true"
rules:
- apiGroups: ["eventing.kyma-project.io"]
  resources: ["subscriptions"]
  verbs: ["create","get", "watch", "list", "delete"]
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: test-event-bus-subs
  labels:
    helm-chart-test: "true"
subjects:
- kind: ServiceAccount
  name: test-event-bus-tester
  namespace: kyma-system
roleRef:
  kind: ClusterRole
  name: test-event-bus-subs
  apiGroup: rbac.authorization.k8s.io
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: test-event-bus-eas
  labels:
    helm-chart-test: "true"
rules:
- apiGroups: ["applicationconnector.kyma-project.io"]
  resources: ["eventactivations"]
  verbs: ["create", "get", "watch", "list", "delete"]
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: test-event-bus-eas
  labels:
    helm-chart-test: "true"
subjects:
- kind: ServiceAccount
  name: test-event-bus-tester
  namespace: kyma-system
roleRef:
  kind: ClusterRole
  name: test-event-bus-eas
  apiGroup: rbac.authorization.k8s.io
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: test-event-bus-k8s
  labels:
    helm-chart-test: "true"
rules:
- apiGroups: ["apps"]
  resources: ["deployments"]
  verbs: ["*"]
- apiGroups: [""]
  resources: ["pods", "pods/status", "services"]
  verbs: ["*"]
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: test-event-bus-k8s
  labels:
    helm-chart-test: "true"
subjects:
- kind: ServiceAccount
  name: test-event-bus-tester
  namespace: kyma-system
roleRef:
  kind: ClusterRole
  name: test-event-bus-k8s
  apiGroup: rbac.authorization.k8s.io
---
# Source: event-bus/charts/publish-knative/templates/metrics-service.yaml
---
apiVersion: v1
kind: Service
metadata:
    name: 'event-bus-publish-knative-metrics-service'
    namespace: kyma-system
    labels:
        app: event-bus-publish-knative
        heritage: "Tiller"
        release: "event-bus"
        chart: publish-knative-0.1.0
spec:
    type: ClusterIP
    ports:
        - name: metrics-port
          port: 9090
    selector:
        app: event-bus-publish-knative
        release: event-bus
---
# Source: event-bus/charts/publish-knative/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: "event-bus-publish"
  namespace: kyma-system
  labels:
    app: publish-knative
    heritage: "Tiller"
    release: "event-bus"
    chart: publish-knative-0.1.0
spec:
  type: ClusterIP
  ports:
    - port: 8080
  selector:
    app: event-bus-publish-knative
    release: event-bus
---
# Source: event-bus/charts/subscription-controller-knative/templates/metrics-service.yaml
---
apiVersion: v1
kind: Service
metadata:
  name: event-bus-subscription-controller-knative-metrics-service
  namespace: kyma-system
  labels:
    app: subscription-controller-knative
    heritage: "Tiller"
    release: "event-bus"
    chart: subscription-controller-knative-0.1.0
spec:
  type: ClusterIP
  ports:
  - name: metrics-port
    port: 9090
  selector:
    app: event-bus-subscription-controller-knative
    release: event-bus
---
# Source: event-bus/charts/subscription-controller-knative/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: event-bus-subscription-controller-knative
  namespace: kyma-system
  labels:
    app: subscription-controller-knative
    heritage: "Tiller"
    release: "event-bus"
    chart: subscription-controller-knative-0.1.0
spec:
  type: ClusterIP
  ports:
    - port: 8080
  selector:
    app: event-bus-subscription-controller-knative
    release: event-bus
---
# Source: event-bus/charts/publish-knative/templates/deployment.yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: event-bus-publish-knative
  namespace: kyma-system
  labels:
    app: publish-knative
    heritage: "Tiller"
    release: "event-bus"
    chart: publish-knative-0.1.0
spec:
  replicas: 1
  selector:
    matchLabels:
      app: event-bus-publish-knative
      release: event-bus
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 0
  template:
    metadata:
      annotations:
        sidecar.istio.io/inject: "true"
      labels:
        app: event-bus-publish-knative
        release: event-bus
        kyma-grafana: enabled
        kyma-alerts: enabled
    spec:
      containers:
        - name: publish-knative
          image: "shazra/event-bus-publish-knative:latest"
          imagePullPolicy: IfNotPresent
          args:
            - --port=8080
            - --max_requests=16
            - --max_request_size=65536
            - --max_channel_name_length=33
            - --trace_api_url=http://zipkin.kyma-system:9411/api/v1/spans
            - --trace_service_name=event-publish-knative-service
            - --trace_operation_name=publish-the-event
            - --max_source_id_length=253
            - --max_event_type_length=253
            - --max_event_type_version_length=4
            - --monitoring_port=9090
          ports:
            - name: http
              containerPort: 8080
          livenessProbe:
            exec:
              command:
              - curl
              - -f
              - http://localhost:8080/v1/status/ready
            initialDelaySeconds: 60
            timeoutSeconds: 10
          resources:
            limits:
              memory: 32M
      serviceAccount:  event-bus-publish
---
# Source: event-bus/charts/subscription-controller-knative/templates/deployment.yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: event-bus-subscription-controller-knative
  namespace: kyma-system
  labels:
    app: subscription-controller-knative
    heritage: "Tiller"
    release: "event-bus"
    chart: subscription-controller-knative-0.1.0
spec:
  replicas: 1
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 0
  template:
    metadata:
      annotations:
        sidecar.istio.io/inject: "false"
      labels:
        app: event-bus-subscription-controller-knative
        release: event-bus
        kyma-grafana: enabled
        kyma-alerts: enabled
    spec:
      containers:
        - name: subscription-controller-knative
          image: "shazra/event-bus-subscription-controller-knative"
          imagePullPolicy: IfNotPresent
          args:
            - --port=8080
            - --resyncPeriod=10s
            - --channelTimeout=10s
          ports:
            - name: http
              containerPort: 8080
          livenessProbe:
            httpGet:
              path: /v1/status/live
              port: http
            initialDelaySeconds: 60
            periodSeconds: 5
          readinessProbe:
            httpGet:
              path: /v1/status/ready
              port: http
            initialDelaySeconds: 60
            periodSeconds: 5
          resources:
            limits:
              memory: 32M
      serviceAccount:  event-bus-subscription-controller-knative-sa
EOF
```

### Install a subscriber/subscription/eventactivation
```bash
cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: event-email-service
  labels:
    example: event-subscription-service
    app: event-email-service
spec:
  type: ClusterIP
  ports:
    - port: 3000
      protocol: TCP
      name: http
      targetPort: 8080
  selector:
    app: event-email-service
---
apiVersion: apps/v1beta2
kind: Deployment
metadata:
  name: event-email-service
  labels:
    app: event-email-service
    example: event-subscription-service
spec:
  replicas: 1
  selector:
    matchLabels:
      app: event-email-service
  template:
    metadata:
      labels:
        app: event-email-service
        example: event-subscription-service
    spec:
      containers:
        - name: event-email-service
          image: shazra/kn-helloworld-go
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 8080
          resources:
            requests:
              cpu: "100m"
            limits:
              memory: "64M"
---
apiVersion: applicationconnector.kyma-project.io/v1alpha1
kind: EventActivation
metadata:
  name: example-event-activation
  labels:
    example: event-subscription-service
spec:
  sourceId: external-application
---
apiVersion: eventing.kyma-project.io/v1alpha1
kind: Subscription
metadata:
  name: example-subscription
  labels:
    example: event-subscription-service
spec:
  endpoint: http://event-email-service.default:3000/v1/events/register
  include_subscription_name_header: true
  event_type: test-event-bus
  event_type_version: v1
  source_id: external-application
EOF
```

### Test EventBus
```bash
# forward requests to the publish app from localhost
kubectl port-forward -n kyma-system $(kubectl get pods -n kyma-system --selector=app=event-bus-publish-knative -o=jsonpath='{.items[0].metadata.name}') 8080:8080

# watch EventBus apps logs
kubectl logs -f -n kyma-system $(kubectl get pods -n kyma-system --selector=app=event-bus-publish-knative -o=jsonpath='{.items[0].metadata.name}') publish-knative

# watch subscription-controller logs
kubectl logs -f -n kyma-system $(kubectl get pods -n kyma-system --selector=app=event-bus-subscription-controller-knative -o=jsonpath='{.items[0].metadata.name}')  subscription-controller-knative

# watch subscriber logs
kubectl logs -f $(kubectl get pods --selector=app=event-email-service -o=jsonpath='{.items[0].metadata.name}')

# watch natss-ch-dispatcher logs
kubectl logs -f -n knative-eventing $(kubectl get pods -n knative-eventing --selector='messaging.knative.dev/channel=natss-channel,messaging.knative.dev/role=dispatcher' -o=jsonpath='{.items[0].metadata.name}') dispatcher

### Check the created Knative subscription
kubectl get subscriptions.eventing.knative.dev -n kyma-system -o yaml

### Check the created messaing channel
kubectl get channels.messaging.knative.dev -n kyma-system

### Check the created natsschannels.messaging.knative.dev
kubectl get natsschannels.messaging.knative.dev -n kyma-system

# publish messages
while true; do \
  curl -i \
      -H "Content-Type: application/json" \
      -X POST http://localhost:8080/v1/events \
      -d '{"source-id": "external-application", "event-type": "test-event-bus", "event-type-version": "v1", "event-time": "2018-11-02T22:08:41+00:00", "data": {"event":{"customer":{"customerID": "'$(date +%s)'", "uid": "rick.sanchez@mail.com"}}}}'; \
  sleep 0.3s; \
done
```

Final proposal
---------------
- Upgrade knative-eventing charts to version 0.8 in Kyma
  - Make sure to include the NatssChannelController and other related resources from [here](https://github.com/knative/eventing/releases/download/v0.8.0/natss.yaml)
- Configure the default channel config to use `NatssChannel`

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: default-ch-webhook
  namespace: knative-eventing
data:
  default-ch-config: |
    clusterDefault:
      apiVersion: messaging.knative.dev/v1alpha1
      kind: NatssChannel 
```

- Subscription controller
  - Watch for Kyma subscriptions(subscriptions.eventing.kyma-project.io)
  - On creation of a new Kyma subscription, follow the create flow:
    - Create a Knative Subscription(aleady happening)
    - Create a Knative Messaging Channel(channels.messaging.knative.dev)
      > The naming convention(using `GetChannelName()`) can be the same as the current one
      > Under the hood a corresponding NatssChannel(natsschannels.messaging.knative.dev) gets created
  - On update of a Kyma subscription, follow the update flow:
    - Delete and create the Knative Subscription(aleady happening)
    - Delete and create the Knative Messaging Channel(channels.messaging.knative.dev)
  - On deletion of an updated Kyma subscription, follow the update flow:
    - Delete the Knative Subscription(aleady happening)
    - Delete the Knative Messaging Channel(channels.messaging.knative.dev)
  
  Bonus:
  - Change the logic for determining the status for Kyma Subscription CR:
    - Watch the status for Knative Subscription and channel.messaging.knative.dev along with EventActivation to determine the readiness of the Kyma subscription 


- Publish service
  - Find the NatssChannel(natsschannels.messaging.knative.dev) based on the similar naming convention
  - Extract the `.status.address.url` from the NatssChannel CR to further POST the event 
  - Do HTTP POST the event to the extracted URL(already happening)
