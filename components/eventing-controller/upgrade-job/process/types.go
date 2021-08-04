package process

import (
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/deployment"
	eventmesh "github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/event-mesh"
	eventingbackend "github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/eventing-backend"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/secret"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/subscription"
)

// Clients consists of clients for resources
type Clients struct {
	Deployment       deployment.Client
	Subscription 	 subscription.Client
	EventingBackend  eventingbackend.Client
	Secret 			 secret.Client
	EventMesh 		 eventmesh.Client
}

// State consists of upgrade-job process state
type State struct {
	Subscriptions          *eventingv1alpha1.SubscriptionList
	FilteredSubscriptions  *eventingv1alpha1.SubscriptionList
	IsBebEnabled           bool
}

// int32Ptr converts int to int pointer
func int32Ptr(i int32) *int32 { return &i }