package object

import (
	"reflect"

	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/conversion"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
)

// Semantic can do semantic deep equality checks for API objects. Fields which
// are not relevant for the reconciliation logic are intentionally omitted.
var Semantic = conversion.EqualitiesOrDie(
	channelEqual,
	peerAuthenticationEqual,
	deploymentEqual,
	serviceEqual,
)

// peerAuthenticationEqual asserts the equality of two PeerAuthentication objects.
func peerAuthenticationEqual(p1, p2 *securityv1beta1.PeerAuthentication) bool {
	if p1 == p2 {
		return true
	}
	if p1 == nil || p2 == nil {
		return false
	}

	if !reflect.DeepEqual(p1.Labels, p2.Labels) {
		return false
	}

	if !reflect.DeepEqual(p1.Spec.Selector, p2.Spec.Selector) {
		return false
	}

	if !reflect.DeepEqual(p1.Spec.PortLevelMtls, p2.Spec.PortLevelMtls) {
		return false
	}

	return true
}

// channelEqual asserts the equality of two Channel objects.
func channelEqual(c1, c2 *messagingv1alpha1.Channel) bool {
	if c1 == c2 {
		return true
	}
	if c1 == nil || c2 == nil {
		return false
	}

	if !reflect.DeepEqual(c1.Labels, c2.Labels) {
		return false
	}
	if !reflect.DeepEqual(c1.Annotations, c2.Annotations) {
		return false
	}

	return true
}

// deploymentEqual asserts the equality of two Deployment objects.
func deploymentEqual(s1, s2 *appsv1.Deployment) bool {
	if s1 == s2 {
		return true
	}
	if s1 == nil || s2 == nil {
		return false
	}

	if !reflect.DeepEqual(s1.Labels, s2.Labels) {
		return false
	}
	if !reflect.DeepEqual(s1.Annotations, s2.Annotations) {
		return false
	}

	cst1 := s1.Spec.Template
	cst2 := s2.Spec.Template

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

func serviceEqual(s1, s2 *corev1.Service) bool {
	if s1 == s2 {
		return true
	}
	if s1 == nil || s2 == nil {
		return false
	}

	if !reflect.DeepEqual(s1.Labels, s2.Labels) {
		return false
	}
	if !reflect.DeepEqual(s1.Annotations, s2.Annotations) {
		return false
	}
	sp1, sp2 := s1.Spec.Ports, s2.Spec.Ports
	if len(sp1) != len(sp2) {
		return false
	}
	for i := range sp1 {
		p1, p2 := &sp1[i], &sp2[i]

		if p1.Name != p2.Name ||
			p1.TargetPort != p2.TargetPort ||
			realProto(p1.Protocol) != realProto(p2.Protocol) || p1.Port != p2.Port {
			return false
		}
	}
	if !reflect.DeepEqual(s1.Spec.Selector, s2.Spec.Selector) {
		return false
	}

	spec1, spec2 := s1.Spec, s2.Spec
	if getServiceType(spec1.Type) != getServiceType(spec2.Type) {
		return false
	}

	if spec1.ClusterIP != spec2.ClusterIP {
		return false
	}
	return true
}

func getServiceType(typ corev1.ServiceType) corev1.ServiceType {
	if typ == "" {
		return corev1.ServiceTypeClusterIP
	}
	return typ
}

// podSpecEqual asserts the equality of two PodSpec objects.
func podSpecEqual(ps1, ps2 *corev1.PodSpec) bool {
	if ps1 == ps2 {
		return true
	}
	if ps1 == nil || ps2 == nil {
		return false
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
	if c1.Image != c2.Image {
		return false
	}

	ps1, ps2 := c1.Ports, c2.Ports
	if len(ps1) != len(ps2) {
		return false
	}
	for i := range ps1 {
		p1, p2 := &ps1[i], &ps2[i]

		if p1.Name != p2.Name ||
			p1.ContainerPort != p2.ContainerPort ||
			realProto(p1.Protocol) != realProto(p2.Protocol) {

			return false
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

func envEqual(e1, e2 []corev1.EnvVar) bool {
	if len(e1) != len(e2) {
		return false
	}
EV1:
	for _, ev1 := range e1 {
		for _, ev2 := range e2 {
			if reflect.DeepEqual(ev1, ev2) {
				continue EV1
			}
		}
		return false
	}
	return true
}

// probeEqual asserts the equality of two Probe objects.
func probeEqual(p1, p2 *corev1.Probe) bool {
	if p1 == p2 {
		return true
	}
	if p1 == nil || p2 == nil {
		return false
	}

	isInitialDelaySecondsEqual := p1.InitialDelaySeconds != p2.InitialDelaySeconds
	isTimeoutSecondsEqual := p1.TimeoutSeconds != p2.TimeoutSeconds && p1.TimeoutSeconds != 0 && p2.TimeoutSeconds != 0
	isPeriodSecondsEqual := p1.PeriodSeconds != p2.PeriodSeconds && p1.PeriodSeconds != 0 && p2.PeriodSeconds != 0
	// Knative sets a default when that value is 0
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

// handlerEqual asserts the equality of two Handler objects.
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

// default Protocol is TCP, so we assume empty equals TCP
// https://godoc.org/k8s.io/api/core/v1#ServicePort
func realProto(pr corev1.Protocol) corev1.Protocol {
	if pr == "" {
		return corev1.ProtocolTCP
	}
	return pr
}
