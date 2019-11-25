package testing

import (
	eventingv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"
)

func GenerateSubscriptionName(name string) k8stesting.ReactionFunc {
	return func(action k8stesting.Action) (handled bool, obj runtime.Object, err error) {
		sub := action.(k8stesting.CreateAction).GetObject().(*eventingv1alpha1.Subscription)

		if sub.GenerateName != "" && sub.Name == "" {
			sub = sub.DeepCopy()
			sub.Name = sub.GenerateName + name
			handled = true
		}

		/*
		   FIXME(antoineco): if the object was handled, fake.Invokes() returns
		   immediately and the object never makes it to the tracker.
		   See https://github.com/kubernetes/client-go/issues/500
		   Fixed in client-go v11.0.0 (k8s 1.13) via https://github.com/kubernetes/kubernetes/pull/73601
		*/
		return handled, sub, nil
	}
}
