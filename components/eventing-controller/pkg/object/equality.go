package object

import (
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/conversion"

	hydrav1alpha1 "github.com/ory/hydra-maester/api/v1alpha1"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"
)

// Semantic can do semantic deep equality checks for API objects. Fields which
// are not relevant for the reconciliation logic are intentionally omitted.
var Semantic = conversion.EqualitiesOrDie(
	apiRuleEqual,
	eventingBackendEqual,
	oauth2ClientEqual,
	publisherProxyDeploymentEqual,
	eventingBackendStatusEqual,
)

// apiRuleEqual asserts the equality of two APIRule objects.
func apiRuleEqual(a1, a2 *apigatewayv1alpha1.APIRule) bool {
	if a1 == nil || a2 == nil {
		return false
	}

	if a1 == a2 {
		return true
	}

	if !reflect.DeepEqual(a1.Labels, a2.Labels) {
		return false
	}

	if !reflect.DeepEqual(a1.OwnerReferences, a2.OwnerReferences) {
		return false
	}
	if !reflect.DeepEqual(a1.Spec.Service.Name, a2.Spec.Service.Name) {
		return false
	}
	if !reflect.DeepEqual(a1.Spec.Service.IsExternal, a2.Spec.Service.IsExternal) {
		return false
	}
	if !reflect.DeepEqual(a1.Spec.Service.Port, a2.Spec.Service.Port) {
		return false
	}
	if !reflect.DeepEqual(a1.Spec.Rules, a2.Spec.Rules) {
		return false
	}
	if !reflect.DeepEqual(a1.Spec.Gateway, a2.Spec.Gateway) {
		return false
	}

	return true
}

// eventingBackendEqual asserts the equality of two EventingBackend objects.
func eventingBackendEqual(b1, b2 *eventingv1alpha1.EventingBackend) bool {
	if b1 == nil || b2 == nil {
		return false
	}
	if b1 == b2 {
		return true
	}

	if !reflect.DeepEqual(b1.Labels, b2.Labels) {
		return false
	}

	if !reflect.DeepEqual(b1.Spec, b2.Spec) {
		return false
	}

	return true
}

// oauth2ClientEqual asserts the equality of two OAuth2Client objects for the
// eventing backend.
func oauth2ClientEqual(oa1, oa2 *hydrav1alpha1.OAuth2Client) bool {
	if oa1 == nil || oa2 == nil {
		return false
	}
	if oa1 == oa2 {
		return true
	}

	if !reflect.DeepEqual(oa1.Labels, oa2.Labels) {
		return false
	}

	if !reflect.DeepEqual(oa1.Spec, oa2.Spec) {
		return false
	}

	return false
}

// publisherProxyDeploymentEqual asserts the equality of two Deployment objects
// for event publisher proxy deployments.
func publisherProxyDeploymentEqual(d1, d2 *appsv1.Deployment) bool {
	if d1 == nil || d2 == nil {
		return false
	}
	if d1 == d2 {
		return true
	}

	if !reflect.DeepEqual(d1.Labels, d2.Labels) {
		return false
	}

	cst1 := d1.Spec.Template
	cst2 := d2.Spec.Template

	if !reflect.DeepEqual(cst1.Annotations, cst2.Annotations) {
		return false
	}

	if !reflect.DeepEqual(cst1.Labels, cst2.Labels) {
		return false
	}

	ps1 := &cst1.Spec
	ps2 := &cst2.Spec
	if !podSpecEqual(ps1, ps2) {
		return false
	}

	return true
}

// podSpecEqual asserts the equality of two PodSpec objects.
func podSpecEqual(ps1, ps2 *corev1.PodSpec) bool {
	if ps1 == nil || ps2 == nil {
		return false
	}
	if ps1 == ps2 {
		return true
	}

	cs1, cs2 := ps1.Containers, ps2.Containers
	if len(cs1) != len(cs2) {
		return false
	}
	for i := range cs1 {
		if !containerEqual(&cs1[i], &cs2[i]) {
			return false
		}
	}

	if ps1.ServiceAccountName != ps2.ServiceAccountName {
		return false
	}

	return true
}

// containerEqual asserts the equality of two Container objects.
func containerEqual(c1, c2 *corev1.Container) bool {
	if c1 == nil || c2 == nil {
		return false
	}
	if c1.Image != c2.Image {
		return false
	}

	ps1, ps2 := c1.Ports, c2.Ports
	if len(ps1) != len(ps2) {
		return false
	}
	for i := range ps1 {
		isFound := false
		for j := range ps2 {
			p1, p2 := &ps1[i], &ps2[j]
			if p1.Name == p2.Name &&
				p1.ContainerPort == p2.ContainerPort &&
				realProto(p1.Protocol) == realProto(p2.Protocol) {
				isFound = true
				break
			}
		}
		if !isFound {
			return isFound
		}
	}

	if !envEqual(c1.Env, c2.Env) {
		return false
	}

	if !probeEqual(c1.ReadinessProbe, c2.ReadinessProbe) {
		return false
	}

	return true
}

// envEqual asserts the queality of two core environment slices. It's used
// by containerEqual.
func envEqual(e1, e2 []corev1.EnvVar) bool {
	if len(e1) != len(e2) {
		return false
	}
	isFound := false
	for _, ev1 := range e1 {
		for _, ev2 := range e2 {
			if reflect.DeepEqual(ev1, ev2) {
				isFound = true
				break
			}
		}
		if !isFound {
			break
		}
	}
	return isFound
}

// probeEqual asserts the equality of two Probe objects. It's used by
// containerEqual.
func probeEqual(p1, p2 *corev1.Probe) bool {
	if p1 == nil || p2 == nil {
		return false
	}

	if p1 == p2 {
		return true
	}

	isInitialDelaySecondsEqual := p1.InitialDelaySeconds != p2.InitialDelaySeconds
	isTimeoutSecondsEqual := p1.TimeoutSeconds != p2.TimeoutSeconds && p1.TimeoutSeconds != 0 && p2.TimeoutSeconds != 0
	isPeriodSecondsEqual := p1.PeriodSeconds != p2.PeriodSeconds && p1.PeriodSeconds != 0 && p2.PeriodSeconds != 0
	isSuccessThresholdEqual := p1.SuccessThreshold != p2.SuccessThreshold && p1.SuccessThreshold != 0 && p2.SuccessThreshold != 0
	isFailureThresholdEqual := p1.FailureThreshold != p2.FailureThreshold && p1.FailureThreshold != 0 && p2.FailureThreshold != 0

	if isInitialDelaySecondsEqual || isTimeoutSecondsEqual || isPeriodSecondsEqual || isSuccessThresholdEqual || isFailureThresholdEqual {
		return false
	}

	if !handlerEqual(&p1.Handler, &p2.Handler) {
		return false
	}

	return true
}

// handlerEqual asserts the equality of two Handler objects. It's used
// by probeEqual.
func handlerEqual(h1, h2 *corev1.Handler) bool {
	if h1 == h2 {
		return true
	}
	if h1 == nil || h2 == nil {
		return false
	}

	hg1, hg2 := h1.HTTPGet, h2.HTTPGet
	if hg1 == nil && hg2 != nil {
		return false
	}
	if hg1 != nil {
		if hg2 == nil {
			return false
		}

		if hg1.Path != hg2.Path {
			return false
		}
	}

	return true
}

// realProto ensures the Protocol, which by default is TCP. We assume empty equals TCP.
// https://godoc.org/k8s.io/api/core/v1#ServicePort
func realProto(pr corev1.Protocol) corev1.Protocol {
	if pr == "" {
		return corev1.ProtocolTCP
	}
	return pr
}

// secretEqual asserts the equality of two Secret objects for event publisher proxy deployments.
func secretEqual(b1, b2 *corev1.Secret) bool {
	if b1 == nil || b2 == nil {
		return false
	}

	if b1 == b2 {
		return true
	}

	if !reflect.DeepEqual(b1.Labels, b2.Labels) {
		return false
	}

	if !reflect.DeepEqual(b1.Data, b2.Data) {
		return false
	}

	return true
}

// eventingBackendStatusEqual asserts the equality of of EventingBackendStatus objects.
func eventingBackendStatusEqual(s1, s2 *eventingv1alpha1.EventingBackendStatus) bool {
	if s1 == nil || s2 == nil {
		return false
	}
	if s1 == s2 {
		return true
	}

	if s1.Backend != s2.Backend {
		return false
	}
	if !utils.BoolPtrEqual(s1.SubscriptionControllerReady, s2.SubscriptionControllerReady) {
		return false
	}
	if !utils.BoolPtrEqual(s1.PublisherProxyReady, s2.PublisherProxyReady) {
		return false
	}
	if !utils.BoolPtrEqual(s1.EventingReady, s2.EventingReady) {
		return false
	}
	if s1.BebSecretName != s2.BebSecretName {
		return false
	}
	if s1.BebSecretNamespace != s2.BebSecretNamespace {
		return false
	}

	return true
}
