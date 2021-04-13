package testing

import (
	"reflect"
	"testing"

	securityv1beta1apis "istio.io/api/security/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"

	"github.com/google/go-cmp/cmp"
	istiofake "istio.io/client-go/pkg/clientset/versioned/fake"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	rt "knative.dev/pkg/reconciler/testing"

	eaFake "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/fake"
)

const knBrokerMetricPort = uint32(9090)

// NewFakeClients initializes fake Clientsets with an optional list of API objects.
func NewFakeClients(objs ...runtime.Object) (*istiofake.Clientset, *eaFake.Clientset) {
	ls := NewListers(objs)

	eaCli := eaFake.NewSimpleClientset(ls.GetEAObjects()...)
	istioCli := istiofake.NewSimpleClientset(ls.GetIstioObjects()...)

	return istioCli, eaCli
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

		wantType := reflect.TypeOf(want).String()
		if wantType == "*v1beta1.PeerAuthentication" {
			if _, ok := want.(*securityv1beta1.PeerAuthentication); ok {
				if _, ok := got.(*securityv1beta1.PeerAuthentication); ok {
					if diff := cmp.Diff(want, got, peerAuthenticationEqual()); diff != "" {
						t.Errorf("Unexpected update (-want, +got): %s", diff)
					}
				}
			}
		} else {
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Unexpected update (-want, +got): %s", diff)
			}
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

		wantType := reflect.ValueOf(want).String()
		if wantType == "*v1beta1.PeerAuthentication" {
			if _, ok := want.(*securityv1beta1.PeerAuthentication); ok {
				if _, ok := got.(*securityv1beta1.PeerAuthentication); ok {
					if diff := cmp.Diff(want, got, peerAuthenticationEqual()); diff != "" {
						t.Errorf("Unexpected update (-want, +got): %s", diff)
					}
				}
			}
		} else {
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Unexpected update (-want, +got): %s", diff)
			}
		}
	}
	if got, want := len(a.Updates), len(expect); got > want {
		for _, extra := range a.Updates[want:] {
			t.Errorf("Extra update: %#v", extra.GetObject())
		}
	}
}

// peerAuthenticationEqual asserts the equality of two istio PeerAuthentication objects.
func peerAuthenticationEqual() cmp.Option {
	alwaysEqual := cmp.Comparer(func(_, _ interface{}) bool { return true })
	return cmp.FilterValues(func(p1, p2 *securityv1beta1.PeerAuthentication) bool {

		if !reflect.DeepEqual(p1.Spec.Selector.MatchLabels, p2.Spec.Selector.MatchLabels) {
			return false
		}

		if !reflect.DeepEqual(p1.Labels, p2.Labels) {
			return false
		}

		if !reflect.DeepEqual(p1.Annotations, p2.Annotations) {
			return false
		}

		var mtlsP1, mtlsP2 *securityv1beta1apis.PeerAuthentication_MutualTLS
		if mP1, ok := p1.Spec.PortLevelMtls[knBrokerMetricPort]; ok {
			mtlsP1 = mP1
		}
		if mP2, ok := p2.Spec.PortLevelMtls[knBrokerMetricPort]; ok {
			mtlsP2 = mP2
		}
		if mtlsP1 == nil || mtlsP2 == nil {
			// t.Logf("Invalid PortLevelMtls for PeerAuthentications: want:%v got:%v", mtlsP1, mtlsP2)
			return false
		}
		if mtlsP1.Mode.String() != mtlsP2.Mode.String() {
			// t.Logf("Invalid PeerAuthentication (-want, +got): -%s +%s", mtlsP1.Mode.String(), mtlsP2.Mode.String())
			return false
		}
		return true
	}, alwaysEqual)
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
