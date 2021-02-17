package testing

import (
	"context"
	"encoding/json"
	"testing"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	fakeeventingclient "knative.dev/eventing/pkg/client/injection/client/fake"
	fakelegacyclient "knative.dev/eventing/pkg/legacyclient/injection/client/fake"
	fakekubeclient "knative.dev/pkg/client/injection/kube/client/fake"
	"knative.dev/pkg/controller"
	fakedynamicclient "knative.dev/pkg/injection/clients/dynamicclient/fake"
	"knative.dev/pkg/logging"
	logtesting "knative.dev/pkg/logging/testing"
	rt "knative.dev/pkg/reconciler/testing"

	fakesourcesclient "github.com/kyma-project/kyma/components/event-sources/client/generated/injection/client/fake"
	fakeistioclient "github.com/kyma-project/kyma/components/event-sources/client/generated/injection/istio/client/fake"
)

const (
	// maxEventBufferSize is the estimated max number of event
	// notifications that can be buffered during reconciliation.
	maxEventBufferSize = 10
)

// Ctor functions create a k8s controller with given params.
type Ctor func(*testing.T, context.Context, *Listers) controller.Reconciler

// MakeFactory creates a reconciler factory with fake clients and controller
// created by the given Ctor.
func MakeFactory(ctor Ctor) rt.Factory {
	return func(t *testing.T, r *rt.TableRow) (controller.Reconciler, rt.ActionRecorderList, rt.EventList, *rt.FakeStatsReporter) {
		scheme := NewScheme()

		ls := NewListers(scheme, r.Objects)

		ctx := context.Background()
		ctx = logging.WithLogger(ctx, logtesting.TestLogger(t))

		ctx, sourcesClient := fakesourcesclient.With(ctx, ls.GetSourcesObjects()...)
		ctx, eventingClient := fakeeventingclient.With(ctx, ls.GetEventingObjects()...)
		ctx, legacyClient := fakelegacyclient.With(ctx, ls.GetLegacyObjects()...)
		ctx, istioClient := fakeistioclient.With(ctx, ls.GetIstioObjects()...)
		// the sink URI resolver lists/watches objects using the dynamic client
		ctx, _ = fakedynamicclient.With(ctx, scheme,
			ToUnstructured(t, scheme, ls.GetEventingObjects())...)
		// also inject fake k8s client which is accessed by reconciler.Base
		k8sobjects := ls.GetCoreObjects()
		k8sobjects = append(k8sobjects, ls.GetAppsObjects()...)
		ctx, kubeClient := fakekubeclient.With(ctx, k8sobjects...)

		eventRecorder := record.NewFakeRecorder(maxEventBufferSize)
		ctx = controller.WithEventRecorder(ctx, eventRecorder)

		// set up Controller from fakes
		c := ctor(t, ctx, &ls)

		actionRecorderList := rt.ActionRecorderList{sourcesClient, eventingClient, istioClient, legacyClient, kubeClient}
		eventList := rt.EventList{Recorder: eventRecorder}
		statsReporter := &rt.FakeStatsReporter{}

		return c, actionRecorderList, eventList, statsReporter
	}
}

// ToUnstructured takes a list of k8s resources and converts them to
// Unstructured objects.
// We must pass objects as Unstructured to the dynamic client fake, or it won't
// handle them properly.
func ToUnstructured(t *testing.T, sch *runtime.Scheme, objs []runtime.Object) (us []runtime.Object) {
	for _, obj := range objs {
		obj = obj.DeepCopyObject() // Don't mess with the primary copy
		// Determine and set the TypeMeta for this object based on our test scheme.
		gvks, _, err := sch.ObjectKinds(obj)
		if err != nil {
			t.Fatalf("Unable to determine kind for type: %v", err)
		}
		apiv, k := gvks[0].ToAPIVersionAndKind()
		ta, err := meta.TypeAccessor(obj)
		if err != nil {
			t.Fatalf("Unable to create type accessor: %v", err)
		}
		ta.SetAPIVersion(apiv)
		ta.SetKind(k)

		b, err := json.Marshal(obj)
		if err != nil {
			t.Fatalf("Unable to marshal: %v", err)
		}
		u := &unstructured.Unstructured{}
		if err := json.Unmarshal(b, u); err != nil {
			t.Fatalf("Unable to unmarshal: %v", err)
		}
		us = append(us, u)
	}
	return
}
