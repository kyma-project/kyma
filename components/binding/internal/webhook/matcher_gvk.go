package webhook

import (
	"fmt"
	"github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

func EqualGVK(a metav1.GroupVersionKind, b schema.GroupVersionKind) bool {
	return a.Kind == b.Kind && a.Version == b.Version && a.Group == b.Group
}

// matchKinds returns error if given obj GVK is not equal to the reqKind GVK
func MatchKinds(obj runtime.Object, reqKind metav1.GroupVersionKind) error {
	err := v1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		return err
	}
	gvk, err := apiutil.GVKForObject(obj, scheme.Scheme)
	if err != nil {
		return err
	}

	if !EqualGVK(reqKind, gvk) {
		return fmt.Errorf("type mismatch: want: %s got: %s", gvk, reqKind)
	}
	return nil
}
