## [POC] Run EventBus on top of Knative [GCPPubSub](https://github.com/google/knative-gcp) CRD

### Export GKE cluster properties
```bash
export CLUSTER_NAME="sayan-event-bus-gcp" \
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

# Apply the second time as it some CRDs which don't get registered right away hence creation of CRs fail.
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

### Create a default configmap for Channel
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
      apiVersion: messaging.cloud.run/v1alpha1
      kind: Channel
EOF
```

### Change RBAC for event-controller ServiceAccount
```bash
cat << EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  annotations:
  labels:
    eventing.knative.dev/release: v0.8.0
  name: knative-eventing-controller
rules:
- apiGroups:
  - ""
  resources:
  - namespaces
  - secrets
  - configmaps
  - services
  - events
  - serviceaccounts
  verbs:
  - get
  - list
  - create
  - update
  - delete
  - patch
  - watch
- apiGroups:
  - apps
  resources:
  - deployments
  verbs:
  - get
  - list
  - create
  - update
  - delete
  - patch
  - watch
- apiGroups:
  - rbac.authorization.k8s.io
  resources:
  - rolebindings
  verbs:
  - get
  - list
  - create
  - update
  - delete
  - patch
  - watch
- apiGroups:
  - eventing.knative.dev
  resources:
  - brokers
  - brokers/status
  - channels
  - channels/status
  - clusterchannelprovisioners
  - clusterchannelprovisioners/status
  - subscriptions
  - subscriptions/status
  - triggers
  - triggers/status
  - eventtypes
  - eventtypes/status
  verbs:
  - get
  - list
  - create
  - update
  - delete
  - patch
  - watch
- apiGroups:
  - eventing.knative.dev
  resources:
  - brokers/finalizers
  - triggers/finalizers
  verbs:
  - update
- apiGroups:
  - messaging.knative.dev
  resources:
  - sequences
  - sequences/status
  - channels
  - channels/status
  - choices
  - choices/status
  verbs:
  - get
  - list
  - create
  - update
  - delete
  - patch
  - watch
- apiGroups:
  - messaging.knative.dev
  resources:
  - sequences/finalizers
  - choices/finalizers
  verbs:
  - update
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - messaging.cloud.run
  resources:
    - channels
  verbs:
    - "*"
EOF
```

> Troubleshooting: The Service Account and the eventing-controller need to be recreated

### Deletion
```
kubectl delete sa -n knative-eventing eventing-controller
kubectl delete deploy -n knative-eventing eventing-controller
```

### Recreation of serviceaccount and eventing-controller(to reflect new RBAC rules)
```
### Create serviceaccount
kubectl create sa -n knative-eventing eventing-controller

### Create deployment for eventing-controller
cat <<EOF | kubectl apply -f -
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    eventing.knative.dev/release: v0.8.0
  name: eventing-controller
  namespace: knative-eventing
spec:
  progressDeadlineSeconds: 600
  replicas: 1
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app: eventing-controller
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      annotations:
        sidecar.istio.io/inject: "false"
      creationTimestamp: null
      labels:
        app: eventing-controller
        eventing.knative.dev/release: v0.8.0
    spec:
      containers:
      - env:
        - name: SYSTEM_NAMESPACE
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
        - name: CONFIG_LOGGING_NAME
          value: config-logging
        - name: CONFIG_OBSERVABILITY_NAME
          value: config-observability
        - name: METRICS_DOMAIN
          value: knative.dev/eventing
        - name: BROKER_INGRESS_IMAGE
          value: gcr.io/knative-releases/github.com/knative/eventing/cmd/broker/ingress@sha256:f029d8d2cb65e31853f8a82394524e45b40958a457450b06f2533e491ffae436
        - name: BROKER_INGRESS_SERVICE_ACCOUNT
          value: eventing-broker-ingress
        - name: BROKER_FILTER_IMAGE
          value: gcr.io/knative-releases/github.com/knative/eventing/cmd/broker/filter@sha256:4d68a946700f184d594de9ebcd9dbbdde7539595fed3e0c3d401d5ee7e2010b6
        - name: BROKER_FILTER_SERVICE_ACCOUNT
          value: eventing-broker-filter
        image: gcr.io/knative-releases/github.com/knative/eventing/cmd/controller@sha256:65d747cf69f3093aec07e4d0a71fab2b16383e990328afb61ff85d62aeb2fa33
        imagePullPolicy: IfNotPresent
        name: eventing-controller
        ports:
        - containerPort: 9090
          name: metrics
          protocol: TCP
        resources: {}
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: FallbackToLogsOnError
        volumeMounts:
        - mountPath: /etc/config-logging
          name: config-logging
      dnsPolicy: ClusterFirst
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      serviceAccount: eventing-controller
      serviceAccountName: eventing-controller
      terminationGracePeriodSeconds: 30
      volumes:
      - configMap:
          defaultMode: 420
          name: config-logging
        name: config-logging
status: {}
EOF
```


### Install EventBus CRD
```bash
# install required CRDs
kubectl apply \
   -f https://raw.githubusercontent.com/kyma-project/kyma/release-1.3/resources/cluster-essentials/templates/event-activation.crd.yaml \
   -f https://raw.githubusercontent.com/kyma-project/kyma/release-1.3/resources/cluster-essentials/templates/eventing-subscription.crd.yaml
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
          image: "shazra/event-bus-publish-knative:1.3"
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
          image: "shazra/event-bus-subscription-controller-knative:1.3"
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

### Create a GCP user/role/key and a corresponding k8s secret

#### Set up your GCP project

```bash
export PROJECT_ID=kyma-project
export KO_DOCKER_REPO=“${dockerhub-username}”
```

#### Install GCPPubSub controller/webhook

Follow the steps mentioned in [here](https://github.com/google/knative-gcp/tree/master/docs/install)

#### Enable the Cloud Pub/Sub API on your project

```bash
gcloud services enable pubsub.googleapis.com
```

#### Create a new service account named cloudrunevents-pullsub with the following command:
> Don't panic if this account is already created

```bash
gcloud iam service-accounts create cloudrunevents-pullsub
```

#### Give that Service Account the Pub/Sub Editor role on your Google Cloud project:

```bash
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member=serviceAccount:cloudrunevents-pullsub@$PROJECT_ID.iam.gserviceaccount.com \
  --role roles/pubsub.editor
```

#### Download a new JSON private key for that Service Account. **Be sure not to check this key into source control!**

```bash
gcloud iam service-accounts keys create cloudrunevents-pullsub.json \
  --iam-account=cloudrunevents-pullsub@$PROJECT_ID.iam.gserviceaccount.com
```

#### Create a secret on the kubernetes cluster with the downloaded key:(used ns is `kyma-system` as the knative subscriptions are created there)

```bash
kubectl --namespace kyma-system create secret generic google-cloud-key --from-file=key.json=cloudrunevents-pullsub.json
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
  event_type: test1-event-bus
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

### Check the created Knative subscription
kubectl get subscriptions.eventing.knative.dev -n kyma-system -o yaml

### Check the created messaging channel
kubectl get channels.messaging.knative.dev -n kyma-system

### Check the created GCPPubSub channel
kubectl get channels.messaging.cloud.run -n kyma-system

# publish messages
while true; do \
  curl -i \
      -H "Content-Type: application/json" \
      -X POST http://localhost:8080/v1/events \
      -d '{"source-id": "external-application", "event-type": "test1-event-bus", "event-type-version": "v1", "event-time": "2018-11-02T22:08:41+00:00", "data": {"event":{"customer":{"customerID": "'$(date +%s)'", "uid": "rick.sanchez@mail.com"}}}}'; \
  sleep 0.3s; \
done
```


### Cleanup
```
gcloud container clusters delete $CLUSTER_NAME --zone $GCP_ZONE
```