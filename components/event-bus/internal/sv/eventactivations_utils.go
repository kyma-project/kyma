package ea

import (
	"log"

	subApis "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma.cx/v1alpha1"
	eaclientset "github.com/kyma-project/kyma/components/event-bus/generated/ea/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// check if an event activation custom resource exists having the same "namespace" as the "sub" namespace and the same "Source"
func checkEventActivationForSubscription(eaClient *eaclientset.Clientset, subObj *subApis.Subscription) bool {
	subNamespace := subObj.GetNamespace()
	subSource := subObj.SubscriptionSpec.Source

	eaList, err := eaClient.RemoteenvironmentV1alpha1().EventActivations(subNamespace).List(metav1.ListOptions{})
	if err != nil {
		log.Printf("Error: List Event Activation call failed for the subscription:\n    %v;\n    Error:%v\n", subObj, err)
		return false
	}
	for _, e := range eaList.Items {
		if subSource.SourceEnvironment == e.Source.Environment &&
			subSource.SourceNamespace == e.Source.Namespace &&
			subSource.SourceType == e.Source.Type {
			return true
		}
	}
	return false
}
