package eventactivation

import (
	"fmt"
	controllertesting "github.com/knative/eventing/pkg/controller/testing"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/event-bus/internal/ea/apis/applicationconnector.kyma-project.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func init() {
	// Add types to scheme
	if err := eventingv1alpha1.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}
}

//var eventActivation = makeNewEventActivation("default", "my-event-activation")

func makeNewEventActivation(namespace string, name string) *eventingv1alpha1.EventActivation {
	eas := eventingv1alpha1.EventActivationSpec{
		DisplayName: "display_name",
		SourceID:    "my_source_ID",
	}
	return &eventingv1alpha1.EventActivation{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EventActivation",
			APIVersion: eventingv1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta:          om(namespace, name),
		EventActivationSpec: eas,
	}
}

func addEventActivationFinalizer(ea *eventingv1alpha1.EventActivation, finalizer string) *eventingv1alpha1.EventActivation {
	ea.ObjectMeta.Finalizers = append(ea.ObjectMeta.Finalizers, finalizer)
	return ea
}

func markedToBeDeletedEventActivation(ea *eventingv1alpha1.EventActivation) *eventingv1alpha1.EventActivation {
	deletedTime := metav1.Now().Rfc3339Copy()
	ea.DeletionTimestamp = &deletedTime
	return ea
}

var testCases = []controllertesting.TestCase{
	{
		Name: "new event activation",
		InitialState: []runtime.Object{
			makeNewEventActivation("default", "my-event-activation"),
		},
		ReconcileKey: fmt.Sprintf("%s/%s", "default", "my-event-activation"),
		WantResult:   reconcile.Result{},
		WantPresent: []runtime.Object{
			addEventActivationFinalizer(
				makeNewEventActivation(
					"default", "my-event-activation"), TestEventActivationFinalizerName),
		},
		Scheme: scheme.Scheme,

		IgnoreTimes: true,
	},
	{
		Name: "Marked to be deleted event activation",
		InitialState: []runtime.Object{
			markedToBeDeletedEventActivation(
				addEventActivationFinalizer(
					makeNewEventActivation("default", "my-event-activation"),
					TestEventActivationFinalizerName)),
		},
		ReconcileKey: fmt.Sprintf("%s/%s", "default", "my-event-activation"),
		WantResult:   reconcile.Result{},
		WantPresent: []runtime.Object{
			markedToBeDeletedEventActivation(
				makeNewEventActivation("default", "my-event-activation")),
		},
		Scheme:      scheme.Scheme,
		IgnoreTimes: true,
	},
}

func TestAllCases(t *testing.T) {
	recorder := record.NewBroadcaster().NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})
	for _, tc := range testCases {
		c := tc.GetClient()
		r := &ReconcileEventActivation{
			Client:   c,
			recorder: recorder,
		}
		t.Logf("Running test %s", tc.Name)
		t.Run(tc.Name, tc.Runner(t, r, c))
	}
}

func TestInjectClient(t *testing.T) {
	println("TestInjectClient()")
	r := &ReconcileEventActivation{}
	orig := r.Client
	n := fake.NewFakeClient()
	if orig == n {
		t.Errorf("Original and new clients are identical: %v", orig)
	}
	err := r.InjectClient(&n)
	if err != nil {
		t.Errorf("Unexpected error injecting the client: %v", err)
	}
	if n != r.Client {
		t.Errorf("Unexpected client. Expected: '%v'. Actual: '%v'", n, r.Client)
	}
}

func om(namespace, name string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Namespace: namespace,
		Name:      name,
		SelfLink:  fmt.Sprintf("/apis/eventing/v1alpha1/namespaces/%s/object/%s", namespace, name),
	}
}
