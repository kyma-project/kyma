/*
Copyright 2019 The Kyma Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package testing

import (
	"context"
	"testing"

	"k8s.io/client-go/tools/record"

	fakeeventingclient "knative.dev/eventing/pkg/client/injection/client/fake"
	fakekubeclient "knative.dev/pkg/client/injection/kube/client/fake"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	fakedynamicclient "knative.dev/pkg/injection/clients/dynamicclient/fake"
	"knative.dev/pkg/logging"
	logtesting "knative.dev/pkg/logging/testing"
	rt "knative.dev/pkg/reconciler/testing"
	fakeservingclient "knative.dev/serving/pkg/client/injection/client/fake"

	fakeclient "github.com/kyma-project/kyma/components/event-sources/client/generated/injection/client/fake"
)

const (
	// maxEventBufferSize is the estimated max number of event
	// notifications that can be buffered during reconciliation.
	maxEventBufferSize = 10
)

// Ctor functions create a k8s controller with given params.
type Ctor func(context.Context, *Listers, configmap.Watcher) controller.Reconciler

// MakeFactory creates a reconciler factory with fake clients and controller
// created by the given Ctor.
func MakeFactory(ctor Ctor) rt.Factory {
	return func(t *testing.T, r *rt.TableRow) (controller.Reconciler, rt.ActionRecorderList, rt.EventList, *rt.FakeStatsReporter) {
		ls := NewListers(r.Objects)

		ctx := r.Ctx
		if ctx == nil {
			ctx = context.Background()
		}

		ctx = logging.WithLogger(ctx, logtesting.TestLogger(t))

		ctx, servingClient := fakeservingclient.With(ctx, ls.GetServingObjects()...)
		ctx, sourcesClient := fakeclient.With(ctx, ls.GetSourcesObjects()...)
		// also inject fake clients accessed by reconciler.Base
		ctx, _ = fakekubeclient.With(ctx)
		ctx, _ = fakeeventingclient.With(ctx)
		ctx, _ = fakedynamicclient.With(ctx, NewScheme())

		// set up Controller from fakes
		c := ctor(ctx, &ls, configmap.NewStaticWatcher())

		eventRecorder := record.NewFakeRecorder(maxEventBufferSize)
		ctx = controller.WithEventRecorder(ctx, eventRecorder)

		actionRecorderList := rt.ActionRecorderList{servingClient, sourcesClient}
		eventList := rt.EventList{Recorder: eventRecorder}
		statsReporter := &rt.FakeStatsReporter{}

		return c, actionRecorderList, eventList, statsReporter
	}
}
