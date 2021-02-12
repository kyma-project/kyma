package fake

import (
	"context"
	"log"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"

	applicationv1alpha1 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application"
)

func NewApplicationListerOrDie(ctx context.Context, app *applicationv1alpha1.Application) *application.Lister {
	scheme := setupSchemeOrDie()
	unstructuredMap := mapToUnstructuredOrDie(app)
	appUnstructured := &unstructured.Unstructured{Object: unstructuredMap}
	appUnstructured.SetGroupVersionKind(application.GroupVersionKind())
	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme, appUnstructured)
	return application.NewLister(ctx, dynamicClient)
}

func setupSchemeOrDie() *runtime.Scheme {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		log.Fatalf("Failed to setup scheme with error: %v", err)
	}
	if err := applicationv1alpha1.AddToScheme(scheme); err != nil {
		log.Fatalf("Failed to setup scheme with error: %v", err)
	}
	return scheme
}

func mapToUnstructuredOrDie(app *applicationv1alpha1.Application) map[string]interface{} {
	unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(app)
	if err != nil {
		log.Fatalf("Failed to map application to unstruchtured object with error: %v", err)
	}
	return unstructuredMap
}
