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

package httpsource

import (
	"context"
	"reflect"

	pkgerrors "github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"

	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	messagingclientv1alpha1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/messaging/v1alpha1"
	messaginglistersv1alpha1 "knative.dev/eventing/pkg/client/listers/messaging/v1alpha1"
	"knative.dev/eventing/pkg/reconciler"
	apisv1alpha1 "knative.dev/pkg/apis/v1alpha1"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/resolver"
	servingv1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"
	servingclientv1alpha1 "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	servinglistersv1alpha1 "knative.dev/serving/pkg/client/listers/serving/v1alpha1"

	sourcesv1alpha1 "github.com/kyma-project/kyma/components/event-sources/apis/sources/v1alpha1"
	sourcesclientv1alpha1 "github.com/kyma-project/kyma/components/event-sources/client/generated/clientset/internalclientset/typed/sources/v1alpha1"
	sourceslistersv1alpha1 "github.com/kyma-project/kyma/components/event-sources/client/generated/lister/sources/v1alpha1"
	"github.com/kyma-project/kyma/components/event-sources/reconciler/errors"
	"github.com/kyma-project/kyma/components/event-sources/reconciler/objects"
)

// Reconciler reconciles HTTPSource resources.
type Reconciler struct {
	// wrapper for core controller components (clients, logger, ...)
	*reconciler.Base

	// adapter properties
	adapterImage string

	// listers index properties about resources
	httpsourceLister sourceslistersv1alpha1.HTTPSourceLister
	ksvcLister       servinglistersv1alpha1.ServiceLister
	chLister         messaginglistersv1alpha1.ChannelLister

	// clients allow interactions with API objects
	sourcesClient   sourcesclientv1alpha1.SourcesV1alpha1Interface
	servingClient   servingclientv1alpha1.ServingV1alpha1Interface
	messagingClient messagingclientv1alpha1.MessagingV1alpha1Interface

	// URI resolver for sink destinations
	sinkResolver *resolver.URIResolver
}

// Reconcile compares the actual state of a HTTPSource object referenced by key
// with its desired state, and attempts to converge the two.
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	src, err := httpSourceByKey(key, r.httpsourceLister)
	if err != nil {
		return errors.Handle(err, ctx, "Failed to get object from local store")
	}

	currentKsvc, err := r.getOrCreateKnService(src)
	if err != nil {
		return err
	}

	currentCh, err := r.getOrCreateChannel(src)
	if err != nil {
		return err
	}

	desiredKsvc := r.makeKnService(src, currentKsvc)
	currentKsvc, err = r.syncKnService(currentKsvc, desiredKsvc)
	if err != nil {
		return pkgerrors.Wrap(err, "failed to synchronize Knative Service")
	}

	return r.syncStatus(src, currentCh, currentKsvc)
}

// httpSourceByKey retrieves a HTTPSource object from a lister by ns/name key.
func httpSourceByKey(key string, l sourceslistersv1alpha1.HTTPSourceLister) (*sourcesv1alpha1.HTTPSource, error) {
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return nil, controller.NewPermanentError(pkgerrors.Wrap(err, "invalid object key"))
	}

	src, err := l.HTTPSources(ns).Get(name)
	switch {
	case apierrors.IsNotFound(err):
		return nil, errors.NewSkippable(pkgerrors.Wrap(err, "object no longer exists"))
	case err != nil:
		return nil, err
	}

	return src, nil
}

// getOrCreateKnService returns the existing Knative Service for a given
// HTTPSource, or creates it if it is missing.
func (r *Reconciler) getOrCreateKnService(src *sourcesv1alpha1.HTTPSource) (*servingv1alpha1.Service, error) {
	ksvc, err := r.ksvcLister.Services(src.Namespace).Get(src.Name)
	switch {
	case apierrors.IsNotFound(err):
		ksvc, err = r.servingClient.Services(src.Namespace).Create(r.makeKnService(src))
		if err != nil {
			return nil, pkgerrors.Wrap(err, "failed to create Knative Service")
		}
	case err != nil:
		return nil, pkgerrors.Wrap(err, "failed to get Knative Service from cache")
	}
	return ksvc, nil
}

// getOrCreateChannel returns the existing Channel for a given HTTPSource, or
// creates it if it is missing.
func (r *Reconciler) getOrCreateChannel(src *sourcesv1alpha1.HTTPSource) (*messagingv1alpha1.Channel, error) {
	ch, err := r.chLister.Channels(src.Namespace).Get(src.Name)
	switch {
	case apierrors.IsNotFound(err):
		ch, err = r.messagingClient.Channels(src.Namespace).Create(r.makeChannel(src))
		if err != nil {
			return nil, pkgerrors.Wrap(err, "failed to create Channel")
		}
	case err != nil:
		return nil, pkgerrors.Wrap(err, "failed to get Channel from cache")
	}
	return ch, nil
}

// makeKnService returns the desired Knative Service object for a given
// HTTPSource. An optional Knative Service can be passed as parameter, in which
// case some of its attributes are used to generate the desired state.
func (r *Reconciler) makeKnService(src *sourcesv1alpha1.HTTPSource,
	currentKsvc ...*servingv1alpha1.Service) *servingv1alpha1.Service {

	var ksvc *servingv1alpha1.Service
	if len(currentKsvc) == 1 {
		ksvc = currentKsvc[0]
	}
	return objects.NewService(src.Namespace, src.Name,
		objects.WithExistingService(ksvc),
		objects.WithContainerImage(r.adapterImage),
		objects.WithServiceControllerRef(src.ToOwner()),
	)
}

// makeChannel returns the desired Channel object for a given HTTPSource.
func (r *Reconciler) makeChannel(src *sourcesv1alpha1.HTTPSource) *messagingv1alpha1.Channel {
	return objects.NewChannel(src.Namespace, src.Name,
		objects.WithChannelControllerRef(src.ToOwner()),
	)
}

// syncKnService synchronizes the desired state of a Knative Service against
// its current state in the running cluster.
func (r *Reconciler) syncKnService(currentKsvc, desiredKsvc *servingv1alpha1.Service) (*servingv1alpha1.Service, error) {
	if objects.Semantic.DeepEqual(currentKsvc, desiredKsvc) {
		return currentKsvc, nil
	}
	return r.servingClient.Services(currentKsvc.Namespace).Update(desiredKsvc)
}

// syncStatus ensures the status of a given HTTPSource is up-to-date.
func (r *Reconciler) syncStatus(src *sourcesv1alpha1.HTTPSource,
	ch *messagingv1alpha1.Channel, ksvc *servingv1alpha1.Service) error {

	currentStatus := &src.Status
	expectedStatus := r.computeStatus(src, ch, ksvc)

	if reflect.DeepEqual(currentStatus, expectedStatus) {
		return nil
	}

	src = &sourcesv1alpha1.HTTPSource{
		ObjectMeta: src.ObjectMeta,
		Status:     *expectedStatus,
	}

	_, err := r.sourcesClient.HTTPSources(src.Namespace).UpdateStatus(src)
	return err
}

// computeStatus returns the expected status of a given HTTPSource.
func (r *Reconciler) computeStatus(src *sourcesv1alpha1.HTTPSource, ch *messagingv1alpha1.Channel,
	ksvc *servingv1alpha1.Service) *sourcesv1alpha1.HTTPSourceStatus {

	status := src.Status.DeepCopy()
	status.InitializeConditions()

	sinkURI, err := r.sinkResolver.URIFromDestination(channelAsDestination(ch), src)
	if err != nil {
		status.MarkNoSink()
		return status
	}
	status.MarkSink(sinkURI)

	status.PropagateServiceReady(ksvc)

	return status
}

// channelAsDestination returns a Destination representation of the given
// Channel.
func channelAsDestination(ch *messagingv1alpha1.Channel) apisv1alpha1.Destination {
	gvk := ch.GetGroupVersionKind()

	return apisv1alpha1.Destination{
		Ref: &corev1.ObjectReference{
			APIVersion: gvk.GroupVersion().String(),
			Kind:       gvk.Kind,
			Namespace:  ch.Namespace,
			Name:       ch.Name,
		},
	}
}
