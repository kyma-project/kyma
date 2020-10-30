package object

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	securityv1beta1apis "istio.io/api/security/v1beta1"
	"istio.io/api/type/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	tTargetPort = 9092
)

func TestNewPeerAuthentication(t *testing.T) {
	applicationLabels := map[string]string{"app": "foo"}
	peerAuthentication := NewPeerAuthentication(tNs, tName,
		WithSelectorSpec(applicationLabels),
		WithPermissiveMode(tTargetPort),
	)

	expectPeerAuthentication := &securityv1beta1.PeerAuthentication{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: tNs,
			Name:      tName,
		},
		Spec: securityv1beta1apis.PeerAuthentication{
			Selector: &v1beta1.WorkloadSelector{
				MatchLabels: applicationLabels,
			},
			PortLevelMtls: map[uint32]*securityv1beta1apis.PeerAuthentication_MutualTLS{
				tTargetPort: {
					Mode: securityv1beta1apis.PeerAuthentication_MutualTLS_PERMISSIVE,
				},
			},
		},
	}

	if d := cmp.Diff(expectPeerAuthentication, peerAuthentication); d != "" {
		t.Errorf("Unexpected diff: (-:expect, +:got) %s", d)
	}
}

func TestApplyExistingPolicyAttributes(t *testing.T) {
	existingPolicy := &securityv1beta1.PeerAuthentication{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       tNs,
			Name:            tName,
			ResourceVersion: "100",
		},
	}

	desiredPolicy := NewPeerAuthentication(tNs, tName)

	ApplyExistingPeerAuthenticationAttributes(existingPolicy, desiredPolicy)
	expectedPolicy := &securityv1beta1.PeerAuthentication{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       tNs,
			Name:            tName,
			ResourceVersion: "100",
		},
	}

	if d := cmp.Diff(desiredPolicy, expectedPolicy); d != "" {
		t.Errorf("Unexpected diff: (-:expect, +:got) %s", d)
	}
}
