package fake

import (
	"log"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"

	"github.com/kyma-project/kyma/components/nats-operator/pkg/client/natscluster"
	"github.com/nats-io/nats-operator/pkg/apis/nats/v1alpha2"
)

// NewClientOrDie TODO ...
func NewClientOrDie(natsClusterList *v1alpha2.NatsClusterList) *natscluster.Client {
	scheme := setupSchemeOrDie()
	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme, natsClusterList)
	return natscluster.NewClient(dynamicClient)
}

// setupSchemeOrDie TODO ...
func setupSchemeOrDie() *runtime.Scheme {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		log.Fatalf("Failed to setup scheme with error: %v", err)
	}
	if err := v1alpha2.AddToScheme(scheme); err != nil {
		log.Fatalf("Failed to setup scheme with error: %v", err)
	}
	return scheme
}
