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

package mqttsource

import (
	"context"
	"reflect"

	pkgerrors "github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"

	"knative.dev/eventing/pkg/reconciler"
	apisv1alpha1 "knative.dev/pkg/apis/v1alpha1"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/resolver"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	servingclientv1 "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1"
	servinglistersv1 "knative.dev/serving/pkg/client/listers/serving/v1"

	"github.com/kyma-project/kyma/components/event-sources/apis/sources/v1alpha1"
	sourcesclientv1alpha1 "github.com/kyma-project/kyma/components/event-sources/client/generated/clientset/internalclientset/typed/sources/v1alpha1"
	sourceslistersv1alpha1 "github.com/kyma-project/kyma/components/event-sources/client/generated/lister/sources/v1alpha1"
	"github.com/kyma-project/kyma/components/event-sources/reconciler/errors"
	"github.com/kyma-project/kyma/components/event-sources/reconciler/objects"
)

// Reconciler reconciles MQTTSource resources.
type Reconciler struct {
	// wrapper for core controller components (clients, logger, ...)
	*reconciler.Base

	// adapter properties
	adapterImage string

	// listers index properties about resources
	mqttsourceLister sourceslistersv1alpha1.MQTTSourceLister
	ksvcLister       servinglistersv1.ServiceLister

	// clients allow interactions with API objects
	sourcesClient sourcesclientv1alpha1.SourcesV1alpha1Interface
	servingClient servingclientv1.ServingV1Interface

	// URI resolver for sink destinations
	sinkResolver *resolver.URIResolver
}

// Reconcile compares the actual state of a MQTTSource object referenced by key
// with its desired state, and attempts to converge the two.
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	mqttSrc, err := mqttSourceByKey(key, r.mqttsourceLister)
	if err != nil {
		return errors.Handle(err, ctx, "Failed to get object from local store")
	}

	currentKsvc, err := r.getOrCreateKnService(mqttSrc)
	if err != nil {
		return err
	}

	desiredKsvc := r.makeKnService(mqttSrc, currentKsvc)
	currentKsvc, err = r.syncKnService(currentKsvc, desiredKsvc)
	if err != nil {
		return pkgerrors.Wrap(err, "failed to synchronize Knative Service")
	}

	return r.syncStatus(mqttSrc, currentKsvc)
}

// mqttSourceByKey retrieves a MQTTSource object from a lister by ns/name key.
func mqttSourceByKey(key string, l sourceslistersv1alpha1.MQTTSourceLister) (*v1alpha1.MQTTSource, error) {
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return nil, controller.NewPermanentError(pkgerrors.Wrap(err, "invalid object key"))
	}

	mqttSrc, err := l.MQTTSources(ns).Get(name)
	switch {
	case apierrors.IsNotFound(err):
		return nil, errors.NewSkippable(pkgerrors.Wrap(err, "object no longer exists"))
	case err != nil:
		return nil, err
	}

	return mqttSrc, nil
}

// getOrCreateKnService returns the existing Knative Service for a given
// MQTTSource, or creates it if it is missing.
func (r *Reconciler) getOrCreateKnService(mqttSrc *v1alpha1.MQTTSource) (*servingv1.Service, error) {
	ksvc, err := r.ksvcLister.Services(mqttSrc.Namespace).Get(mqttSrc.Name)
	switch {
	case apierrors.IsNotFound(err):
		ksvc, err = r.servingClient.Services(mqttSrc.Namespace).Create(r.makeKnService(mqttSrc))
		if err != nil {
			return nil, pkgerrors.Wrap(err, "failed to create Knative Service")
		}
	case err != nil:
		return nil, pkgerrors.Wrap(err, "failed to get Knative Service from cache")
	}
	return ksvc, nil
}

// makeKnService returns the desired Knative Service object for a given
// MQTTSource. An optional Knative Service can be passed as parameter, in which
// case some of its attributes are used to generate the desired state.
func (r *Reconciler) makeKnService(mqttSrc *v1alpha1.MQTTSource, currentKsvc ...*servingv1.Service) *servingv1.Service {
	opts := []objects.ServiceOption{
		objects.WithContainerImage(r.adapterImage),
		objects.WithControllerRef(mqttSrc.ToOwner()),
	}
	if len(currentKsvc) == 1 {
		opts = append(opts, objects.WithExisting(currentKsvc[0]))
	}
	return objects.NewService(mqttSrc.Namespace, mqttSrc.Name, opts...)
}

// syncKnService synchronizes the desired state of a Knative Service against
// its current state in the running cluster.
func (r *Reconciler) syncKnService(currentKsvc, desiredKsvc *servingv1.Service) (*servingv1.Service, error) {
	if objects.Semantic.DeepEqual(currentKsvc, desiredKsvc) {
		return currentKsvc, nil
	}
	return r.servingClient.Services(currentKsvc.Namespace).Update(desiredKsvc)
}

// syncStatus ensures the status of a given MQTTSource is up-to-date.
func (r *Reconciler) syncStatus(mqttSrc *v1alpha1.MQTTSource, currentKsvc *servingv1.Service) error {
	statusCpy := mqttSrc.Status.DeepCopy()
	statusCpy.InitializeConditions()

	sinkURI, err := r.sinkResolver.URIFromDestination(getSinkDest(), mqttSrc)
	if err != nil {
		statusCpy.MarkNoSink("NotFound", "The sink does not exist")
		return err
	}
	statusCpy.MarkSink(sinkURI)

	statusCpy.PropagateServiceReady(currentKsvc)

	if !reflect.DeepEqual(statusCpy, &mqttSrc.Status) {
		mqttSrc = &v1alpha1.MQTTSource{
			ObjectMeta: mqttSrc.ObjectMeta,
			Status:     *statusCpy,
		}

		_, err = r.sourcesClient.MQTTSources(mqttSrc.Namespace).UpdateStatus(mqttSrc)
		if err != nil {
			return err
		}
	}

	return nil
}

// TODO: find a way to not have this hardcoded
func getSinkDest() apisv1alpha1.Destination {
	return apisv1alpha1.Destination{
		Ref: &corev1.ObjectReference{
			APIVersion: "v1",
			Kind:       "Service",
			Namespace:  "kyma-system",
			Name:       "event-publish-service",
		},
	}
}
