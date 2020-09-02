package object

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/pkg/errors"
	securityv1beta1apis "istio.io/api/security/v1beta1"
	"istio.io/api/type/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
)

const (
	fixtureChannelPath            = "../../test/fixtures/channel.json"
	fixturePeerAuthenticationPath = "../../test/fixtures/peerAuthentication.json"
	fixtureDeploymentPath         = "../../test/fixtures/deployment.json"
	fixtureServicePath            = "../../test/fixtures/service.json"
)

var fixtureChannel *messagingv1alpha1.Channel
var fixtureDeployment *appsv1.Deployment
var fixturePeerAuthentication *securityv1beta1.PeerAuthentication
var fixtureService *corev1.Service

func TestMain(m *testing.M) {
	var err error

	fixtureChannel = &messagingv1alpha1.Channel{}
	if err = loadFixture(fixtureChannelPath, fixtureChannel); err != nil {
		panic(errors.Wrap(err, "loading Channel from fixtures"))
	}

	fixturePeerAuthentication = &securityv1beta1.PeerAuthentication{}
	if err = loadFixture(fixturePeerAuthenticationPath, fixturePeerAuthentication); err != nil {
		panic(errors.Wrap(err, "loading PeerAuthentication from fixtures"))
	}

	fixtureDeployment = &appsv1.Deployment{}
	if err = loadFixture(fixtureDeploymentPath, fixtureDeployment); err != nil {
		panic(errors.Wrap(err, "loading Deployment from fixtures"))
	}

	fixtureService = &corev1.Service{}
	if err = loadFixture(fixtureServicePath, fixtureService); err != nil {
		panic(errors.Wrap(err, "loading Deployment from fixtures"))
	}

	os.Exit(m.Run())
}

func loadFixture(file string, obj metav1.Object) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, obj); err != nil {
		return err
	}

	return nil
}

func TestChannelEqual(t *testing.T) {
	ch := fixtureChannel

	if !channelEqual(nil, nil) {
		t.Error("Two nil elements should be equal")
	}

	testCases := map[string]struct {
		prep   func() *messagingv1alpha1.Channel
		expect bool
	}{
		"not equal when one element is nil": {
			func() *messagingv1alpha1.Channel {
				return nil
			},
			false,
		},
		"not equal when labels differ": {
			func() *messagingv1alpha1.Channel {
				chCopy := ch.DeepCopy()
				chCopy.Labels["foo"] += "test"
				return chCopy
			},
			false,
		},
		"not equal when annotations differ": {
			func() *messagingv1alpha1.Channel {
				chCopy := ch.DeepCopy()
				chCopy.Annotations["foo"] += "test"
				return chCopy
			},
			false,
		},
		"equal when other fields differ": {
			func() *messagingv1alpha1.Channel {
				chCopy := ch.DeepCopy()

				// metadata
				lbls := chCopy.Labels
				anns := chCopy.Annotations

				m := &chCopy.ObjectMeta
				m.Reset()
				m.Labels = lbls
				m.Annotations = anns

				// spec
				sp := &chCopy.Spec
				*sp = messagingv1alpha1.ChannelSpec{} // reset

				// status
				st := &chCopy.Status
				*st = messagingv1alpha1.ChannelStatus{} // reset

				return chCopy
			},
			true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			testCh := tc.prep()
			if channelEqual(testCh, ch) != tc.expect {
				t.Errorf("Expected output to be %t", tc.expect)
			}
		})
	}
}

func TestDeploymentEqual(t *testing.T) {
	deployment := fixtureDeployment

	if !deploymentEqual(nil, nil) {
		t.Error("Two nil elements should be equal")
	}

	testCases := map[string]struct {
		prep   func() *appsv1.Deployment
		expect bool
	}{
		"not equal when one element is nil": {
			func() *appsv1.Deployment {
				return nil
			},
			false,
		},
		"not equal when labels differ": {
			func() *appsv1.Deployment {
				deploymentCopy := deployment.DeepCopy()
				deploymentCopy.Labels["foo"] += "test"
				return deploymentCopy
			},
			false,
		},
		"not equal when annotations differ": {
			func() *appsv1.Deployment {
				deploymentCopy := deployment.DeepCopy()
				deploymentCopy.Annotations["foo"] += "test"
				return deploymentCopy
			},
			false,
		},
		"not equal when template annotations differ": {
			func() *appsv1.Deployment {
				deploymentCopy := deployment.DeepCopy()
				deployment.Spec.Template.ObjectMeta.Annotations["foo"] += "test"
				return deploymentCopy
			},
			false,
		},
		"not equal when template labels differ": {
			func() *appsv1.Deployment {
				deploymentCopy := deployment.DeepCopy()
				deployment.Spec.Template.ObjectMeta.Labels["foo"] += "test"
				return deploymentCopy
			},
			false,
		},
		"equal when other fields differ": {
			func() *appsv1.Deployment {
				deploymentCopy := deployment.DeepCopy()

				// metadata
				lbls := deploymentCopy.Labels
				anns := deploymentCopy.Annotations

				m := &deploymentCopy.ObjectMeta
				m.Reset()
				m.Labels = lbls
				m.Annotations = anns

				// spec
				sp := &deploymentCopy.Spec

				tplAnns := sp.Template.ObjectMeta.Annotations
				tplLbls := sp.Template.ObjectMeta.Labels
				ps := sp.Template.Spec

				*sp = appsv1.DeploymentSpec{} // reset
				sp.Template = corev1.PodTemplateSpec{}
				sp.Template.ObjectMeta.Annotations = tplAnns
				sp.Template.ObjectMeta.Labels = tplLbls
				sp.Template.Spec = ps

				// status
				st := &deploymentCopy.Status
				*st = appsv1.DeploymentStatus{} // reset

				return deploymentCopy
			},
			true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			testDeployment := tc.prep()
			if deploymentEqual(testDeployment, deployment) != tc.expect {
				t.Errorf("Expected output to be %t", tc.expect)
			}
		})
	}
}

func TestServiceEqual(t *testing.T) {
	svc := fixtureService

	if !serviceEqual(nil, nil) {
		t.Error("Two nil elements should be equal")
	}

	testCases := map[string]struct {
		prep   func() *corev1.Service
		expect bool
	}{
		"not equal when one element is nil": {
			func() *corev1.Service {
				return nil
			},
			false,
		},
		"not equal when labels differ": {
			func() *corev1.Service {
				svcCopy := svc.DeepCopy()
				svcCopy.Labels["foo"] += "test"
				return svcCopy
			},
			false,
		},
		"not equal when annotations differ": {
			func() *corev1.Service {
				svcCopy := svc.DeepCopy()
				svcCopy.Annotations = map[string]string{"a": "b"}
				return svcCopy
			},
			false,
		},
		"not equal when type differs": {
			func() *corev1.Service {
				svcCopy := svc.DeepCopy()
				svcCopy.Spec.Type = corev1.ServiceTypeExternalName
				return svcCopy
			},
			false,
		},
		"equal when type gets defaulted": {
			func() *corev1.Service {
				svcCopy := svc.DeepCopy()
				svcCopy.Spec.Type = ""
				return svcCopy
			},
			true,
		},
		"not equal when ClusterIP differs": {
			func() *corev1.Service {
				svcCopy := svc.DeepCopy()
				svcCopy.Spec.ClusterIP = "1.1.1.1"
				return svcCopy
			},
			false,
		},
		"not equal when Ports differs": {
			func() *corev1.Service {
				svcCopy := svc.DeepCopy()
				svcCopy.Spec.Ports = []corev1.ServicePort{}
				return svcCopy
			},
			false,
		},
		"not equal when selector differs": {
			func() *corev1.Service {
				svcCopy := svc.DeepCopy()
				svcCopy.Spec.Selector = map[string]string{"a": "b"}
				return svcCopy
			},
			false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			testService := tc.prep()
			if serviceEqual(testService, svc) != tc.expect {
				t.Errorf("Expected output to be %t", tc.expect)
			}
		})
	}

}

func TestPodSpecEqual(t *testing.T) {
	ps := &fixtureDeployment.Spec.Template.Spec

	if ps.Containers == nil {
		t.Fatalf("Test requires fixture object %s to have a least one Container", fixtureDeploymentPath)
	}

	if !podSpecEqual(nil, nil) {
		t.Error("Two nil elements should be equal")
	}

	testCases := map[string]struct {
		prep   func() *corev1.PodSpec
		expect bool
	}{
		"not equal when one element is nil": {
			func() *corev1.PodSpec {
				return nil
			},
			false,
		},
		"not equal when ServiceAccount field differs": {
			func() *corev1.PodSpec {
				psCopy := ps.DeepCopy()
				psCopy.ServiceAccountName += "foo"
				return psCopy
			},
			false,
		},
		"not equal when number of containers differs": {
			func() *corev1.PodSpec {
				psCopy := ps.DeepCopy()
				psCopy.Containers = append(psCopy.Containers, corev1.Container{})
				return psCopy
			},
			false,
		},
		"equal when other fields differ": {
			func() *corev1.PodSpec {
				psCopy := ps.DeepCopy()

				containers := psCopy.Containers
				saName := psCopy.ServiceAccountName

				psCopy.Reset()
				psCopy.Containers = containers
				psCopy.ServiceAccountName = saName

				return psCopy
			},
			true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			testSpec := tc.prep()
			if podSpecEqual(testSpec, ps) != tc.expect {
				t.Errorf("Expected output to be %t", tc.expect)
			}
		})
	}
}

func TestContainerEqual(t *testing.T) {
	cs := fixtureDeployment.Spec.Template.Spec.Containers

	if cs == nil {
		t.Fatalf("Test requires fixture object %s to have a least one Container", fixtureDeploymentPath)
	}
	if cs[0].Ports == nil {
		t.Fatalf("Test requires fixture object %s to have a least one ContainerPort", fixtureDeploymentPath)
	}
	if cs[0].Env == nil {
		t.Fatalf("Test requires fixture object %s to have a least one EnvVar", fixtureDeploymentPath)
	}

	c := &cs[0]

	testCases := map[string]struct {
		prep   func() *corev1.Container
		expect bool
	}{
		"not equal when image differs": {
			func() *corev1.Container {
				cCopy := c.DeepCopy()
				cCopy.Image += "test"
				return cCopy
			},
			false,
		},
		"not equal when number of ports differs": {
			func() *corev1.Container {
				cCopy := c.DeepCopy()
				cCopy.Ports = append(cCopy.Ports, corev1.ContainerPort{})
				return cCopy
			},
			false,
		},
		"not equal when number of envvars differs": {
			func() *corev1.Container {
				cCopy := c.DeepCopy()
				cCopy.Env = append(cCopy.Env, corev1.EnvVar{})
				return cCopy
			},
			false,
		},
		"not equal when envvar differs": {
			func() *corev1.Container {
				cCopy := c.DeepCopy()
				cCopy.Env[0].Value += "test"
				return cCopy
			},
			false,
		},
		"equal when other fields differ": {
			func() *corev1.Container {
				cCopy := c.DeepCopy()

				img := cCopy.Image
				ports := cCopy.Ports
				env := cCopy.Env
				probe := cCopy.ReadinessProbe

				cCopy.Reset()
				cCopy.Image = img
				cCopy.Ports = ports
				cCopy.Env = env
				cCopy.ReadinessProbe = probe

				return cCopy
			},
			true,
		},
	}

	for _, fieldName := range []string{
		"Name",
		"ContainerPort",
		"Protocol",
	} {
		testCases[fmt.Sprintf("not equal when Port.%s field differs", fieldName)] = struct {
			prep   func() *corev1.Container
			expect bool
		}{
			func(fieldName string) func() *corev1.Container {
				return func() *corev1.Container {
					cCopy := c.DeepCopy()

					p := &cCopy.Ports[0]
					ptrP := reflect.ValueOf(p)
					pElem := ptrP.Elem()
					f := pElem.FieldByName(fieldName)

					switch f.Kind() {
					case reflect.String:
						f.SetString(f.String() + "test")
					case reflect.Int32:
						f.SetInt(f.Int() + 1)
					}

					return cCopy
				}
			}(fieldName),
			false,
		}
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			testC := tc.prep()
			if containerEqual(testC, c) != tc.expect {
				t.Errorf("Expected output to be %t", tc.expect)
			}
		})
	}
}

func TestEnvEqual(t *testing.T) {

	testCases := map[string]struct {
		a      []corev1.EnvVar
		b      []corev1.EnvVar
		expect bool
	}{
		"equal": {
			[]corev1.EnvVar{
				{
					Name:  "a",
					Value: "a",
				},
				{
					Name:  "b",
					Value: "b",
				},
			},
			[]corev1.EnvVar{
				{
					Name:  "a",
					Value: "a",
				},
				{
					Name:  "b",
					Value: "b",
				},
			},
			true,
		},
		"equal different order": {
			[]corev1.EnvVar{
				{
					Name:  "a",
					Value: "a",
				},
				{
					Name:  "b",
					Value: "b",
				},
			},
			[]corev1.EnvVar{
				{
					Name:  "b",
					Value: "b",
				},
				{
					Name:  "a",
					Value: "a",
				},
			},
			true,
		},
		"different same length": {
			[]corev1.EnvVar{
				{
					Name:  "a",
					Value: "a",
				},
				{
					Name:  "b",
					Value: "b",
				},
			},
			[]corev1.EnvVar{
				{
					Name:  "a",
					Value: "a",
				},
				{
					Name:  "c",
					Value: "c",
				},
			},
			false,
		},
		"different with different length": {
			[]corev1.EnvVar{
				{
					Name:  "a",
					Value: "a",
				},
				{
					Name:  "b",
					Value: "b",
				},
			},
			[]corev1.EnvVar{
				{
					Name:  "a",
					Value: "a",
				},
			},
			false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if envEqual(tc.a, tc.b) != tc.expect {
				t.Errorf("Expected output to be %t", tc.expect)
			}
		})
	}
}
func TestHandlerEqual(t *testing.T) {
	cs := fixtureDeployment.Spec.Template.Spec.Containers

	if cs == nil || cs[0].ReadinessProbe == nil || cs[0].ReadinessProbe.HTTPGet == nil {
		t.Fatalf("Test requires fixture object %s to have a configured HTTPGet Handler", fixtureDeploymentPath)
	}

	httpH := &cs[0].ReadinessProbe.Handler

	testCases := map[string]struct {
		prep   func() *corev1.Handler
		expect bool
	}{
		"not equal when one element is nil": {
			func() *corev1.Handler {
				return nil
			},
			false,
		},
		"not equal when http Handler Path field differs": {
			func() *corev1.Handler {
				hCopy := httpH.DeepCopy()
				hCopy.HTTPGet.Path += "test"
				return hCopy
			},
			false,
		},
		"equal when other http Handler fields differ": {
			func() *corev1.Handler {
				hCopy := httpH.DeepCopy()
				hCopy.HTTPGet.Port = intstr.FromInt(hCopy.HTTPGet.Port.IntValue() + 1)
				hCopy.HTTPGet.Host += "test"
				hCopy.HTTPGet.Scheme += "test"
				hCopy.HTTPGet.HTTPHeaders = append(hCopy.HTTPGet.HTTPHeaders, corev1.HTTPHeader{})
				return hCopy
			},
			true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			testH := tc.prep()
			if handlerEqual(testH, httpH) != tc.expect {
				t.Errorf("Expected output to be %t", tc.expect)
			}
		})
	}
}

func TestPeerAuthenticationEqual(t *testing.T) {
	peerAuthentication := fixturePeerAuthentication
	testCases := map[string]struct {
		prep   func() *securityv1beta1.PeerAuthentication
		expect bool
	}{
		"not equal when selector differs": {
			prep: func() *securityv1beta1.PeerAuthentication {
				p := peerAuthentication.DeepCopy()
				p.Spec.Selector = &v1beta1.WorkloadSelector{
					MatchLabels: map[string]string{"foo": "bar"},
				}
				return p
			},
			expect: false,
		},
		"not equal when portlevel mtls port differs": {
			prep: func() *securityv1beta1.PeerAuthentication {
				p := peerAuthentication.DeepCopy()
				differentPort := uint32(1234)
				p.Spec.PortLevelMtls = map[uint32]*securityv1beta1apis.PeerAuthentication_MutualTLS{
					differentPort: {
						Mode: securityv1beta1apis.PeerAuthentication_MutualTLS_PERMISSIVE,
					},
				}
				return p
			},
			expect: false,
		},
		"not equal when portlevel mtls authentication differs": {
			prep: func() *securityv1beta1.PeerAuthentication {
				p := peerAuthentication.DeepCopy()
				p.Spec.PortLevelMtls[tTargetPort].Mode = securityv1beta1apis.PeerAuthentication_MutualTLS_STRICT
				return p
			},
			expect: false,
		},
		"equal when labels, mtls and selector are equal but other fields are different": {
			func() *securityv1beta1.PeerAuthentication {
				p := peerAuthentication.DeepCopy()
				p.Annotations = map[string]string{
					"foo": fmt.Sprintf("%s%s", p.Annotations["foo"], "bar"),
				}
				p.Spec.XXX_NoUnkeyedLiteral = struct{}{}
				p.Spec.XXX_unrecognized = nil
				p.Spec.XXX_sizecache = 0
				return p
			},
			true,
		},
		"not equal when labels differ": {
			prep: func() *securityv1beta1.PeerAuthentication {
				p := peerAuthentication.DeepCopy()
				p.Labels = map[string]string{
					"foo": fmt.Sprintf("%s%s", p.Labels["foo"], "bar"),
				}
				return p
			},
			expect: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			testPeerAuthentication := tc.prep()
			if peerAuthenticationEqual(testPeerAuthentication, peerAuthentication) != tc.expect {
				t.Errorf("Expected output to be %t", tc.expect)
			}
		})
	}
}
