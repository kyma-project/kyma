## [POC] Run EventBus on top of Knative NatssChannel CRD

### Export GKE cluster properties
```bash
export CLUSTER_NAME="knative-event-bus" \
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
kubectl apply --filename https://raw.githubusercontent.com/knative/serving/v0.7.0/third_party/istio-1.1.7/istio-crds.yaml &&
curl -L https://raw.githubusercontent.com/knative/serving/v0.7.0/third_party/istio-1.1.7/istio.yaml \
  | sed 's/LoadBalancer/NodePort/' \
  | kubectl apply --filename -

# label the default namespace with istio-injection=enabled.
kubectl label namespace default istio-injection=enabled

# watch pods until is shows STATUS of Running or Completed
kubectl get pods --namespace istio-system -w
```

### Install Knative
```bash
kubectl apply --selector knative.dev/crd-install=true \
   --filename https://github.com/knative/serving/releases/download/v0.7.0/serving.yaml \
   --filename https://github.com/knative/eventing/releases/download/v0.7.0/release.yaml

kubectl apply --filename https://github.com/knative/serving/releases/download/v0.7.0/serving.yaml --selector networking.knative.dev/certificate-provider!=cert-manager \
   --filename https://github.com/knative/eventing/releases/download/v0.7.0/release.yaml
```

### Install NATSS server
```bash
kubectl create namespace natss; \
kubectl apply -n natss -f https://raw.githubusercontent.com/knative/eventing/v0.7.0/contrib/natss/config/broker/natss.yaml
```

### Install NATSS controller and dispatcher
```bash
# export your docker registry and use ko to install NATSS controller and dispatcher
export KO_DOCKER_REPO=marcobebway; \
cd $GOPATH/src/github.com/knative/eventing/contrib/natss/config; \
git fetch; \
git checkout v0.7.0; \
ko apply -f .; \
git checkout -
```

### Checkout the POC branch
```bash
cd $GOPATH/src/github.com/kyma-project/kyma/; \
git remote add marcobebway git@github.com:marcobebway/kyma.git; \
git fetch marcobebway -v; \
git checkout marcobebway/poc-knative-natss-channel-crd
```

### Install EventBus
```bash
# install required CRDs
kubectl apply \
   -f https://raw.githubusercontent.com/kyma-project/kyma/release-1.3/resources/cluster-essentials/templates/event-activation.crd.yaml \
   -f https://raw.githubusercontent.com/kyma-project/kyma/release-1.3/resources/cluster-essentials/templates/eventing-subscription.crd.yaml

# create kyma-system namespace with istio-injection enabled
kubectl create ns kyma-system; \
kubectl label namespace kyma-system istio-injection=enabled

# install EventBus
kubectl apply -f $GOPATH/src/github.com/kyma-project/kyma/components/event-bus/poc/event-bus.yaml -n kyma-system
```

### Create NatssChannel instance inside the kyma-system namespace
```bash
cat << EOF | kubectl apply -f -
apiVersion: messaging.knative.dev/v1alpha1
kind: NatssChannel
metadata:
  name: my-test-channel
  namespace: kyma-system
EOF
```

### Update EventBus ClusterRoles to query NatssChannel
```bash
cat << EOF | kubectl apply -f -
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
      - natsschannels
      - natsschannels/status
    verbs:
      - get
      - list
      - watch
---
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
      - natsschannels
      - natsschannels/status
    verbs:
      - get
      - list
      - watch
EOF
```

### Create new images for the EventBus apps (Optional - can skip and use the images in the next step)
```bash
# export your docker registry to push images
export DOCKER_REGISTRY=marcobebway; \
cd $GOPATH/src/github.com/kyma-project/kyma/components/event-bus/; \
make clean build-image; \
docker tag event-bus-publish-knative $DOCKER_REGISTRY/event-bus-publish-knative:latest; \
docker push $DOCKER_REGISTRY/event-bus-publish-knative; \
docker tag event-bus-subscription-controller-knative $DOCKER_REGISTRY/event-bus-subscription-controller-knative:latest; \
docker push $DOCKER_REGISTRY/event-bus-subscription-controller-knative
```

### Update EventBus apps to use the new docker images
```bash
# update the publish app image using the image below or the one from your docker registry
kubectl edit deployment -n kyma-system event-bus-publish-knative
> image: marcobebway/event-bus-publish-knative:latest
> imagePullPolicy: Always

# update the subscription-controller app image using the image below or the one from your docker registry
kubectl edit deployment -n kyma-system event-bus-subscription-controller-knative
> image: marcobebway/event-bus-subscription-controller-knative:latest
> imagePullPolicy: Always
```

### Test EventBus
```bash
# setup EventBus test resources (EventActivation, Kyma Subscription and an example Subscriber)
kubectl apply -f $GOPATH/src/github.com/kyma-project/kyma/components/event-bus/poc/event-bus-test.yaml

# forward requests to the publish app from localhost
kubectl port-forward -n kyma-system $(kubectl get pods -n kyma-system --selector=app=event-bus-publish-knative -o=jsonpath='{.items[0].metadata.name}') 8080:8080

# watch EventBus apps logs
kubectl logs -f -n kyma-system $(kubectl get pods -n kyma-system --selector=app=event-bus-publish-knative -o=jsonpath='{.items[0].metadata.name}') publish-knative
kubectl logs -f -n kyma-system $(kubectl get pods -n kyma-system --selector=app=event-bus-subscription-controller-knative -o=jsonpath='{.items[0].metadata.name}')  subscription-controller-knative

# watch subscriber logs
kubectl logs -f $(kubectl get pods --selector=app=event-email-service -o=jsonpath='{.items[0].metadata.name}')

# watch natss-ch-dispatcher logs
kubectl logs -f -n knative-eventing $(kubectl get pods -n knative-eventing --selector='messaging.knative.dev/channel=natss-channel,messaging.knative.dev/role=dispatcher' -o=jsonpath='{.items[0].metadata.name}') dispatcher

# publish messages
while true; do \
  curl -i \
      -H "Content-Type: application/json" \
      -X POST http://localhost:8080/v1/events \
      -d '{"source-id": "external-application", "event-type": "test-event-bus", "event-type-version": "v1", "event-time": "2018-11-02T22:08:41+00:00", "data": {"event":{"customer":{"customerID": "'$(date +%s)'", "uid": "rick.sanchez@mail.com"}}}}'; \
  sleep 0.3s; \
done
```

### Check the created Knative subscription
```bash
kubectl get subscriptions.eventing.knative.dev -n kyma-system example-subscription--default -o yaml

# Knative subscription
apiVersion: eventing.knative.dev/v1alpha1
kind: Subscription
metadata:
  annotations:
    eventing.knative.dev/creator: system:serviceaccount:kyma-system:event-bus-subscription-controller-knative-sa
    eventing.knative.dev/lastModifier: system:serviceaccount:kyma-system:event-bus-subscription-controller-knative-sa
  creationTimestamp: 2019-07-23T18:53:52Z
  finalizers:
  - subscription-controller
  generation: 1
  name: example-subscription--default
  namespace: kyma-system
  resourceVersion: "17132"
  selfLink: /apis/eventing.knative.dev/v1alpha1/namespaces/kyma-system/subscriptions/example-subscription--default
  uid: 36daa464-ad7b-11e9-b082-42010a8400ab
spec:
  channel:
    apiVersion: messaging.knative.dev/v1alpha1
    kind: NatssChannel
    name: my-test-channel
  reply: {}
  subscriber:
    uri: http://event-email-service.default:3000/v1/events/register
status:
  conditions:
  - lastTransitionTime: 2019-07-23T18:53:52Z
    status: "True"
    type: AddedToChannel
  - lastTransitionTime: 2019-07-23T18:53:52Z
    status: "True"
    type: ChannelReady
  - lastTransitionTime: 2019-07-23T18:53:52Z
    status: "True"
    type: Ready
  - lastTransitionTime: 2019-07-23T18:53:52Z
    status: "True"
    type: Resolved
  physicalSubscription:
    subscriberURI: http://event-email-service.default:3000/v1/events/register
```
