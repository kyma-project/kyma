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
	"fmt"
	"reflect"

	pkgerrors "github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"

	authenticationv1alpha1 "istio.io/client-go/pkg/apis/authentication/v1alpha1"
	authenticationclientv1alpha1 "istio.io/client-go/pkg/clientset/versioned/typed/authentication/v1alpha1"
	authenticationlistersv1alpha1 "istio.io/client-go/pkg/listers/authentication/v1alpha1"
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
	routeconfig "knative.dev/serving/pkg/reconciler/route/config"

	sourcesv1alpha1 "github.com/kyma-project/kyma/components/event-sources/apis/sources/v1alpha1"
	sourcesclientv1alpha1 "github.com/kyma-project/kyma/components/event-sources/client/generated/clientset/internalclientset/typed/sources/v1alpha1"
	sourceslistersv1alpha1 "github.com/kyma-project/kyma/components/event-sources/client/generated/lister/sources/v1alpha1"
	"github.com/kyma-project/kyma/components/event-sources/reconciler/errors"
	"github.com/kyma-project/kyma/components/event-sources/reconciler/object"
)

// Reconciler reconciles HTTPSource resources.
type Reconciler struct {
	// wrapper for core controller components (clients, logger, ...)
	*reconciler.Base

	// adapter properties
	adapterEnvCfg     *httpAdapterEnvConfig
	adapterMetricsCfg string
	adapterLoggingCfg string

	// listers index properties about resources
	httpsourceLister sourceslistersv1alpha1.HTTPSourceLister
	ksvcLister       servinglistersv1alpha1.ServiceLister
	chLister         messaginglistersv1alpha1.ChannelLister
	policyLister     authenticationlistersv1alpha1.PolicyLister

	// clients allow interactions with API objects
	sourcesClient   sourcesclientv1alpha1.SourcesV1alpha1Interface
	servingClient   servingclientv1alpha1.ServingV1alpha1Interface
	messagingClient messagingclientv1alpha1.MessagingV1alpha1Interface
	policyClient    authenticationclientv1alpha1.AuthenticationV1alpha1Interface

	// URI resolver for sink destinations
	sinkResolver *resolver.URIResolver
}

// Mandatory adapter env vars
const (
	// Common
	// see https://github.com/knative/eventing/blob/release-0.10/pkg/adapter/config.go
	sinkURIEnvVar       = "SINK_URI"
	namespaceEnvVar     = "NAMESPACE"
	metricsConfigEnvVar = "K_METRICS_CONFIG"
	loggingConfigEnvVar = "K_LOGGING_CONFIG"

	// HTTP adapter specific
	eventSourceEnvVar = "EVENT_SOURCE"

	// Private svc suffix of an Ksvc
	privateSvcSuffix = "-private"
)

const adapterHealthEndpoint = "/healthz"

const applicationNameLabelKey = "application-name"

// Reconcile compares the actual state of a HTTPSource object referenced by key
// with its desired state, and attempts to converge the two.
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	src, err := httpSourceByKey(key, r.httpsourceLister)
	if err != nil {
		return errors.Handle(err, ctx, "Failed to get object from local store")
	}

	ch, err := r.reconcileSink(src)
	if err != nil {
		return err
	}

	ksvc, err := r.reconcileAdapter(src, ch)
	if err != nil {
		return err
	}

	policy, err := r.reconcilePolicy(src, ksvc)
	if err != nil {
		return err
	}
	return r.syncStatus(src, ch, ksvc, policy)
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

// reconcileSink reconciles the state of the HTTP adapter's sink (Channel).
func (r *Reconciler) reconcileSink(src *sourcesv1alpha1.HTTPSource) (*messagingv1alpha1.Channel, error) {
	desiredCh := r.makeChannel(src)

	currentCh, err := r.getOrCreateChannel(src, desiredCh)
	if err != nil {
		return nil, err
	}

	object.ApplyExistingChannelAttributes(currentCh, desiredCh)

	currentCh, err = r.syncChannel(src, currentCh, desiredCh)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "failed to synchronize Channel")
	}

	return currentCh, nil
}

// reconcilePolicy reconciles the state of the Policy.
func (r *Reconciler) reconcilePolicy(src *sourcesv1alpha1.HTTPSource, ksvc *servingv1alpha1.Service) (*authenticationv1alpha1.Policy, error) {
	if ksvc == nil {
		r.event(src, failedCreateReason, "Skipping creation of Policy as there is no ksvc yet")
		return nil, nil
	}

	if &ksvc.Status != nil && len(ksvc.Status.ConfigurationStatusFields.LatestCreatedRevisionName) == 0 {
		r.event(src, failedCreateReason, "Skipping creation of Policy as there is no revision yet")
		return nil, nil
	}
	desiredPolicy := r.makePolicy(src, ksvc)
	currentPolicy, err := r.getOrCreatePolicy(src, desiredPolicy)
	if err != nil {
		return nil, err
	}

	object.ApplyExistingPolicyAttributes(currentPolicy, desiredPolicy)

	currentPolicy, err = r.syncPolicy(src, currentPolicy, desiredPolicy)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "failed to synchronize Policy")
	}

	return desiredPolicy, nil
}

// reconcileAdapter reconciles the state of the HTTP adapter.
func (r *Reconciler) reconcileAdapter(src *sourcesv1alpha1.HTTPSource,
	sink *messagingv1alpha1.Channel) (*servingv1alpha1.Service, error) {

	sinkURI, err := getSinkURI(sink, r.sinkResolver, src)
	if err != nil {
		// delay reconciliation until the sink becomes ready
		return nil, nil
	}

	desiredKsvc := r.makeKnService(src, sinkURI, r.adapterLoggingCfg, r.adapterMetricsCfg)

	currentKsvc, err := r.getOrCreateKnService(src, desiredKsvc)
	if err != nil {
		return nil, err
	}

	object.ApplyExistingServiceAttributes(currentKsvc, desiredKsvc)

	currentKsvc, err = r.syncKnService(src, currentKsvc, desiredKsvc)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "failed to synchronize Knative Service")
	}

	return currentKsvc, nil
}

// getOrCreateKnService returns the existing Knative Service for a given
// HTTPSource, or creates it if it is missing.
func (r *Reconciler) getOrCreateKnService(src *sourcesv1alpha1.HTTPSource,
	desiredKsvc *servingv1alpha1.Service) (*servingv1alpha1.Service, error) {

	ksvc, err := r.ksvcLister.Services(src.Namespace).Get(src.Name)
	switch {
	case apierrors.IsNotFound(err):
		ksvc, err = r.servingClient.Services(src.Namespace).Create(desiredKsvc)
		if err != nil {
			r.eventWarn(src, failedCreateReason, "Creation failed for Knative Service %q", desiredKsvc.Name)
			return nil, pkgerrors.Wrap(err, "failed to create Knative Service")
		}
		r.event(src, createReason, "Created Knative Service %q", ksvc.Name)

	case err != nil:
		return nil, pkgerrors.Wrap(err, "failed to get Knative Service from cache")
	}

	return ksvc, nil
}

// getOrCreateChannel returns the existing Channel for a given HTTPSource, or
// creates it if it is missing.
func (r *Reconciler) getOrCreateChannel(src *sourcesv1alpha1.HTTPSource,
	desiredCh *messagingv1alpha1.Channel) (*messagingv1alpha1.Channel, error) {

	ch, err := r.chLister.Channels(src.Namespace).Get(src.Name)
	switch {
	case apierrors.IsNotFound(err):
		ch, err = r.messagingClient.Channels(src.Namespace).Create(desiredCh)
		if err != nil {
			r.eventWarn(src, failedCreateReason, "Creation failed for Channel %q", desiredCh.Name)
			return nil, pkgerrors.Wrap(err, "failed to create Channel")
		}
		r.event(src, createReason, "Created Channel %q", ch.Name)

	case err != nil:
		return nil, pkgerrors.Wrap(err, "failed to get Channel from cache")
	}

	return ch, nil
}

// getOrCreatePolicy returns the existing Policy for a Revision of a KnativeService, or
// creates it if it is missing.
func (r *Reconciler) getOrCreatePolicy(src *sourcesv1alpha1.HTTPSource,
	desiredPolicy *authenticationv1alpha1.Policy) (*authenticationv1alpha1.Policy, error) {
	policy, err := r.policyLister.Policies(src.Namespace).Get(desiredPolicy.Name)
	switch {
	case apierrors.IsNotFound(err):
		policy, err = r.policyClient.Policies(src.Namespace).Create(desiredPolicy)
		if err != nil {
			r.eventWarn(src, failedCreateReason, "Creation failed for Policy %q", desiredPolicy.Name)
			return nil, pkgerrors.Wrap(err, "failed to create Policy")
		}
		r.event(src, createReason, "Created Policy %q", policy.Name)

	case err != nil:
		return nil, pkgerrors.Wrap(err, "failed to get Policy from cache")
	}

	return policy, nil
}

// makeKnService returns the desired Knative Service object for a given
// HTTPSource. An optional Knative Service can be passed as parameter, in which
// case some of its attributes are used to generate the desired state.
func (r *Reconciler) makeKnService(src *sourcesv1alpha1.HTTPSource,
	sinkURI, loggingCfg, metricsCfg string) *servingv1alpha1.Service {

	return object.NewService(src.Namespace, src.Name,
		object.WithImage(r.adapterEnvCfg.Image),
		object.WithPort(r.adapterEnvCfg.Port),
		object.WithMinScale(1),
		object.WithEnvVar(eventSourceEnvVar, src.Spec.Source),
		object.WithEnvVar(sinkURIEnvVar, sinkURI),
		object.WithEnvVar(namespaceEnvVar, src.Namespace),
		object.WithEnvVar(metricsConfigEnvVar, metricsCfg),
		object.WithEnvVar(loggingConfigEnvVar, loggingCfg),
		object.WithProbe(adapterHealthEndpoint),
		object.WithControllerRef(src.ToOwner()),
		object.WithLabel(routeconfig.VisibilityLabelKey, routeconfig.VisibilityClusterLocal),
		object.WithPodLabel(dashboardLabelKey, dashboardLabelValue),
	)
}

// makeChannel returns the desired Channel object for a given HTTPSource.
func (r *Reconciler) makeChannel(src *sourcesv1alpha1.HTTPSource) *messagingv1alpha1.Channel {
	return object.NewChannel(src.Namespace, src.Name,
		object.WithControllerRef(src.ToOwner()),
		object.WithLabel(applicationNameLabelKey, src.Name),
	)
}

// makePolicy returns the desired Policy object for a given Ksvc per HTTPSource.
func (r *Reconciler) makePolicy(src *sourcesv1alpha1.HTTPSource, ksvc *servingv1alpha1.Service) *authenticationv1alpha1.Policy {
	// Using the private k8s svc of a ksvc which has the metrics ports
	name := fmt.Sprintf("%s%s", ksvc.Status.ConfigurationStatusFields.LatestCreatedRevisionName, privateSvcSuffix)
	return object.NewPolicy(src.Namespace, name,
		object.WithControllerRef(src.ToOwner()),
		object.WithLabel(applicationNameLabelKey, src.Name),
		object.WithTarget(name),
	)
}

// syncKnService synchronizes the desired state of a Knative Service against
// its current state in the running cluster.
func (r *Reconciler) syncKnService(src *sourcesv1alpha1.HTTPSource,
	currentKsvc, desiredKsvc *servingv1alpha1.Service) (*servingv1alpha1.Service, error) {

	if object.Semantic.DeepEqual(currentKsvc, desiredKsvc) {
		return currentKsvc, nil
	}

	ksvc, err := r.servingClient.Services(currentKsvc.Namespace).Update(desiredKsvc)
	if err != nil {
		r.eventWarn(src, failedUpdateReason, "Update failed for Knative Service %q", desiredKsvc.Name)
		return nil, err
	}
	r.event(src, updateReason, "Updated Knative Service %q", ksvc.Name)

	return ksvc, nil
}

// syncChannel synchronizes the desired state of a Channel against its current
// state in the running cluster.
func (r *Reconciler) syncChannel(src *sourcesv1alpha1.HTTPSource,
	currentCh, desiredCh *messagingv1alpha1.Channel) (*messagingv1alpha1.Channel, error) {

	if object.Semantic.DeepEqual(currentCh, desiredCh) {
		return currentCh, nil
	}

	ch, err := r.messagingClient.Channels(currentCh.Namespace).Update(desiredCh)
	if err != nil {
		r.eventWarn(src, failedUpdateReason, "Update failed for Channel %q", desiredCh.Name)
		return nil, err
	}
	r.event(src, updateReason, "Updated Channel %q", ch.Name)

	return ch, nil
}

// syncPolicy synchronizes the desired state of a Policy against its current
// state in the running cluster.
func (r *Reconciler) syncPolicy(src *sourcesv1alpha1.HTTPSource,
	currentPolicy, desiredPolicy *authenticationv1alpha1.Policy) (*authenticationv1alpha1.Policy, error) {

	if object.Semantic.DeepEqual(currentPolicy, desiredPolicy) {
		return currentPolicy, nil
	}

	policy, err := r.policyClient.Policies(currentPolicy.Namespace).Update(desiredPolicy)
	if err != nil {
		r.eventWarn(src, failedUpdateReason, "Update failed for Policy %q", desiredPolicy.Name)
		return nil, err
	}
	r.event(src, updateReason, "Updated Policy %q", policy.Name)

	return policy, nil
}

// syncStatus ensures the status of a given HTTPSource is up-to-date.
func (r *Reconciler) syncStatus(src *sourcesv1alpha1.HTTPSource,
	ch *messagingv1alpha1.Channel, ksvc *servingv1alpha1.Service, policy *authenticationv1alpha1.Policy) error {

	currentStatus := &src.Status
	expectedStatus := r.computeStatus(src, ch, ksvc, policy)

	if reflect.DeepEqual(currentStatus, expectedStatus) {
		return nil
	}

	src = &sourcesv1alpha1.HTTPSource{
		ObjectMeta: src.ObjectMeta,
		Status:     *expectedStatus,
		// sending the Spec in a status update is optional, however
		// fake ClientSets apply the same UpdateActionImpl action to
		// the object tracker regardless of the subresource
		Spec: src.Spec,
	}

	_, err := r.sourcesClient.HTTPSources(src.Namespace).UpdateStatus(src)
	return err
}

// computeStatus returns the expected status of a given HTTPSource.
func (r *Reconciler) computeStatus(src *sourcesv1alpha1.HTTPSource, ch *messagingv1alpha1.Channel,
	ksvc *servingv1alpha1.Service, policy *authenticationv1alpha1.Policy) *sourcesv1alpha1.HTTPSourceStatus {

	status := src.Status.DeepCopy()
	status.InitializeConditions()

	sinkURI, err := r.sinkResolver.URIFromDestination(channelAsDestination(ch), src)
	if err != nil {
		status.MarkNoSink()
		return status
	}
	status.MarkSink(sinkURI)
	status.MarkMonitoring(policy)
	if ksvc != nil {
		status.PropagateServiceReady(ksvc)
	}

	return status
}

// getSinkURI returns the URI of a Channel. An error is returned if the URI
// can't be determined which, can also happen when the Channel is not ready.
func getSinkURI(sink *messagingv1alpha1.Channel, r *resolver.URIResolver, parent interface{}) (string, error) {
	sinkURI, err := r.URIFromDestination(channelAsDestination(sink), parent)
	if err != nil {
		return "", pkgerrors.Wrap(err, "failed to read sink URI")
	}
	return sinkURI, nil
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
