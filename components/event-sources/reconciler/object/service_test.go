package object

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/google/go-cmp/cmp"
)

const (
	tSelLabelKey   = "sellabelkey"
	tSelLabelValue = "sellabelvalue"
	tPortName      = "portName"
	tExternalPort  = 80
	tContainerPort = 8080
	tLabelKey      = "labelkey"
	tLabelValue    = "labelvalue"
)

func TestNewService(t *testing.T) {
	service := NewService(tNs, tName,
		WithSelector(tSelLabelKey, tSelLabelValue),
		WithServicePort(tPortName, tExternalPort, tContainerPort),
		WithLabel(tLabelKey, tLabelValue))

	expectedService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: tNs,
			Name:      tName,
			Labels: map[string]string{
				tLabelKey: tLabelValue,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       tPortName,
					Port:       tExternalPort,
					TargetPort: intstr.FromInt(tContainerPort),
				},
			},
			Selector: map[string]string{
				tSelLabelKey: tSelLabelValue,
			},
		},
	}

	if d := cmp.Diff(expectedService, service); d != "" {
		t.Errorf("Unexpected diff: (-:expect, +:got) %s", d)
	}
}
