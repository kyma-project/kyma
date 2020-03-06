package testing

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	eaFake "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/fake"
	istiofake "istio.io/client-go/pkg/clientset/versioned/fake"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	eventingfake "knative.dev/eventing/pkg/client/clientset/versioned/fake"
	rt "knative.dev/pkg/reconciler/testing"
)

// NewFakeClients initializes fake Clientsets with an optional list of API objects.
func NewFakeClients(objs ...runtime.Object) (*eventingfake.Clientset, *k8sfake.Clientset, *istiofake.Clientset, *eaFake.Clientset) {
	ls := NewListers(objs)

	eaCli := eaFake.NewSimpleClientset(ls.GetEAObjects()...)
	evCli := eventingfake.NewSimpleClientset(ls.GetEventingObjects()...)
	k8sCli := k8sfake.NewSimpleClientset(ls.GetKubeObjects()...)
	istioCli := istiofake.NewSimpleClientset(ls.GetIstioObjects()...)

	return evCli, k8sCli, istioCli, eaCli
}

type ActionsAsserter struct {
	rt.Actions
}

func NewActionsAsserter(t *testing.T, clis ...rt.ActionRecorder) *ActionsAsserter {
	t.Helper()

	actionRecorderList := rt.ActionRecorderList(clis)

	actions, err := actionRecorderList.ActionsByVerb()
	if err != nil {
		t.Fatalf("Failed to get clients actions by verb: %s", err)
	}

	return &ActionsAsserter{Actions: actions}
}

func (a *ActionsAsserter) AssertCreates(t *testing.T, expect []runtime.Object) {
	t.Helper()

	for i, want := range expect {
		if i >= len(a.Creates) {
			t.Errorf("Missing create: %#v", want)
			continue
		}

		got := a.Creates[i].GetObject()

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("Unexpected create (-want, +got): %s", diff)
		}
	}
	if got, want := len(a.Creates), len(expect); got > want {
		for _, extra := range a.Creates[want:] {
			t.Errorf("Extra create: %#v", extra.GetObject())
		}
	}
}

func (a *ActionsAsserter) AssertUpdates(t *testing.T, expect []runtime.Object) {
	t.Helper()

	for i, want := range expect {
		if i >= len(a.Updates) {
			t.Errorf("Missing update: %#v", want)
			continue
		}

		got := a.Updates[i].GetObject()

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("Unexpected update (-want, +got): %s", diff)
		}
	}
	if got, want := len(a.Updates), len(expect); got > want {
		for _, extra := range a.Updates[want:] {
			t.Errorf("Extra update: %#v", extra.GetObject())
		}
	}
}

func (a *ActionsAsserter) AssertDeletes(t *testing.T, expect []string) {
	t.Helper()

	for i, want := range expect {
		if i >= len(a.Deletes) {
			t.Errorf("Missing delete: %#v", want)
			continue
		}

		got := a.Deletes[i]

		wantNs, wantName, _ := cache.SplitMetaNamespaceKey(want)
		if got.GetNamespace() != wantNs || got.GetName() != wantName {
			t.Errorf("Unexpected delete: %#v", got)
		}
	}
	if got, want := len(a.Deletes), len(expect); got > want {
		for _, extra := range a.Deletes[want:] {
			t.Errorf("Extra delete: %#v", extra)
		}
	}
}
