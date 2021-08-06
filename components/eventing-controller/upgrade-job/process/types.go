package process

import (
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/configmap"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/deployment"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/eventingbackend"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/eventmesh"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/secret"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/subscription"
)

// Clients consists of clients for resources
type Clients struct {
	Deployment      deployment.Client
	Subscription    subscription.Client
	EventingBackend eventingbackend.Client
	Secret          secret.Client
	ConfigMap       configmap.Client
	EventMesh       eventmesh.Client
}

// State consists of upgrade-job process state
type State struct {
	Subscriptions         *eventingv1alpha1.SubscriptionList
	FilteredSubscriptions *eventingv1alpha1.SubscriptionList
	//BebSubscriptionNameMapper   *handlers.NameMapper
	IsBebEnabled bool
}

// int32Ptr converts int to int pointer
func int32Ptr(i int32) *int32 { return &i }

// PatchDeploymentSpec for defining patch data for updating k8s deployment
type PatchDeploymentSpec struct {
	Spec Spec `json:"spec"`
}

// Spec child node for PatchDeploymentSpec struct
type Spec struct {
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`
}
