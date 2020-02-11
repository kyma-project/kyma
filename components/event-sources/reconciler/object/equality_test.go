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

package object

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	authenticationv1alpha1api "istio.io/api/authentication/v1alpha1"
	authenticationv1alpha1 "istio.io/client-go/pkg/apis/authentication/v1alpha1"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	servingv1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"
)

const (
	fixtureChannelPath = "../../test/fixtures/channel.json"
	fixtureKsvcPath    = "../../test/fixtures/ksvc.json"
	fixturePolicyPath  = "../../test/fixtures/policy.json"
)

var fixtureChannel *messagingv1alpha1.Channel
var fixtureKsvc *servingv1alpha1.Service
var fixturePolicy *authenticationv1alpha1.Policy

func TestMain(m *testing.M) {
	var err error

	fixtureChannel = &messagingv1alpha1.Channel{}
	if err = loadFixture(fixtureChannelPath, fixtureChannel); err != nil {
		panic(errors.Wrap(err, "loading Channel from fixtures"))
	}

	fixtureKsvc = &servingv1alpha1.Service{}
	if err = loadFixture(fixtureKsvcPath, fixtureKsvc); err != nil {
		panic(errors.Wrap(err, "loading Knative Service from fixtures"))
	}

	fixturePolicy = &authenticationv1alpha1.Policy{}
	if err = loadFixture(fixturePolicyPath, fixturePolicy); err != nil {
		panic(errors.Wrap(err, "loading Policy from fixtures"))
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

func TestKsvcEqual(t *testing.T) {
	ksvc := fixtureKsvc

	if !ksvcEqual(nil, nil) {
		t.Error("Two nil elements should be equal")
	}

	testCases := map[string]struct {
		prep   func() *servingv1alpha1.Service
		expect bool
	}{
		"not equal when one element is nil": {
			func() *servingv1alpha1.Service {
				return nil
			},
			false,
		},
		"not equal when labels differ": {
			func() *servingv1alpha1.Service {
				ksvcCopy := ksvc.DeepCopy()
				ksvcCopy.Labels["foo"] += "test"
				return ksvcCopy
			},
			false,
		},
		"not equal when annotations differ": {
			func() *servingv1alpha1.Service {
				ksvcCopy := ksvc.DeepCopy()
				ksvcCopy.Annotations["foo"] += "test"
				return ksvcCopy
			},
			false,
		},
		"not equal when template annotations differ": {
			func() *servingv1alpha1.Service {
				ksvcCopy := ksvc.DeepCopy()
				ksvc.Spec.ConfigurationSpec.Template.ObjectMeta.Annotations["foo"] += "test"
				return ksvcCopy
			},
			false,
		},
		"not equal when template labels differ": {
			func() *servingv1alpha1.Service {
				ksvcCopy := ksvc.DeepCopy()
				ksvc.Spec.ConfigurationSpec.Template.ObjectMeta.Labels["foo"] += "test"
				return ksvcCopy
			},
			false,
		},
		"equal when other fields differ": {
			func() *servingv1alpha1.Service {
				ksvcCopy := ksvc.DeepCopy()

				// metadata
				lbls := ksvcCopy.Labels
				anns := ksvcCopy.Annotations

				m := &ksvcCopy.ObjectMeta
				m.Reset()
				m.Labels = lbls
				m.Annotations = anns

				// spec
				sp := &ksvcCopy.Spec

				tplAnns := sp.ConfigurationSpec.Template.ObjectMeta.Annotations
				tplLbls := sp.ConfigurationSpec.Template.ObjectMeta.Labels
				ps := sp.ConfigurationSpec.Template.Spec.PodSpec

				*sp = servingv1alpha1.ServiceSpec{} // reset
				sp.ConfigurationSpec.Template = &servingv1alpha1.RevisionTemplateSpec{}
				sp.ConfigurationSpec.Template.ObjectMeta.Annotations = tplAnns
				sp.ConfigurationSpec.Template.ObjectMeta.Labels = tplLbls
				sp.ConfigurationSpec.Template.Spec.PodSpec = ps

				// status
				st := &ksvcCopy.Status
				*st = servingv1alpha1.ServiceStatus{} // reset

				return ksvcCopy
			},
			true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			testKsvc := tc.prep()
			if ksvcEqual(testKsvc, ksvc) != tc.expect {
				t.Errorf("Expected output to be %t", tc.expect)
			}
		})
	}
}

func TestPodSpecEqual(t *testing.T) {
	ps := &fixtureKsvc.Spec.ConfigurationSpec.Template.Spec.PodSpec

	if ps.Containers == nil {
		t.Fatalf("Test requires fixture object %s to have a least one Container", fixtureKsvcPath)
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
	cs := fixtureKsvc.Spec.ConfigurationSpec.Template.Spec.PodSpec.Containers

	if cs == nil {
		t.Fatalf("Test requires fixture object %s to have a least one Container", fixtureKsvcPath)
	}
	if cs[0].Ports == nil {
		t.Fatalf("Test requires fixture object %s to have a least one ContainerPort", fixtureKsvcPath)
	}
	if cs[0].Env == nil {
		t.Fatalf("Test requires fixture object %s to have a least one EnvVar", fixtureKsvcPath)
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

func TestProbeEqual(t *testing.T) {
	cs := fixtureKsvc.Spec.ConfigurationSpec.Template.Spec.PodSpec.Containers

	if cs == nil || cs[0].ReadinessProbe == nil {
		t.Fatalf("Test requires fixture object %s to have a configured ReadinessProbe", fixtureKsvcPath)
	}
	if cs[0].ReadinessProbe.SuccessThreshold == 0 {
		t.Fatalf("Test requires fixture object %s ReadinessProbe to have a defined SuccessThreshold", fixtureKsvcPath)
	}

	p := cs[0].ReadinessProbe

	testCases := map[string]struct {
		prep   func() *corev1.Probe
		expect bool
	}{
		"not equal when one element is nil": {
			func() *corev1.Probe {
				return nil
			},
			false,
		},
		"not equal when handlers differ": {
			func() *corev1.Probe {
				pCopy := p.DeepCopy()
				pCopy.Handler.Reset()
				return pCopy
			},
			false,
		},
	}

	for _, fieldName := range []string{
		"InitialDelaySeconds",
		"TimeoutSeconds",
		"PeriodSeconds",
		"SuccessThreshold",
		"FailureThreshold",
	} {
		testCases[fmt.Sprintf("not equal when %s field differs", fieldName)] = struct {
			prep   func() *corev1.Probe
			expect bool
		}{
			func(fieldName string) func() *corev1.Probe {
				return func() *corev1.Probe {
					pCopy := p.DeepCopy()

					ptrProbe := reflect.ValueOf(pCopy)
					probe := ptrProbe.Elem()
					f := probe.FieldByName(fieldName)
					f.SetInt(f.Int() + 1)

					return pCopy
				}
			}(fieldName),
			false,
		}
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			testP := tc.prep()
			if probeEqual(testP, p) != tc.expect {
				t.Errorf("Expected output to be %t", tc.expect)
			}
		})
	}
}

func TestHandlerEqual(t *testing.T) {
	cs := fixtureKsvc.Spec.ConfigurationSpec.Template.Spec.PodSpec.Containers

	if cs == nil || cs[0].ReadinessProbe == nil || cs[0].ReadinessProbe.HTTPGet == nil {
		t.Fatalf("Test requires fixture object %s to have a configured HTTPGet Handler", fixtureKsvcPath)
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

func TestPolicyEqual(t *testing.T) {
	policy := fixturePolicy
	testCases := map[string]struct {
		prep   func() *authenticationv1alpha1.Policy
		expect bool
	}{
		"not equal when targets differ": {
			prep: func() *authenticationv1alpha1.Policy {
				p := policy.DeepCopy()
				p.Spec = authenticationv1alpha1api.Policy{
					Targets: []*authenticationv1alpha1api.TargetSelector{
						{Name: "foo"},
						{Name: "bar"},
					},
				}
				return p
			},
			expect: false,
		},
		"not equal when peers differ": {
			prep: func() *authenticationv1alpha1.Policy {
				p := policy.DeepCopy()
				p.Spec.Peers = []*authenticationv1alpha1api.PeerAuthenticationMethod{
					{
						Params: &authenticationv1alpha1api.PeerAuthenticationMethod_Mtls{
							Mtls: &authenticationv1alpha1api.MutualTls{
								Mode: authenticationv1alpha1api.MutualTls_STRICT,
							},
						},
					},
				}
				return p
			},
			expect: false,
		},
		"equal when labels, targets and peers are equal but other fields are different": {
			func() *authenticationv1alpha1.Policy {
				p := policy.DeepCopy()
				p.Annotations = map[string]string{
					"foo": fmt.Sprintf("%s%s", p.Annotations["foo"], "bar"),
				}
				p.Spec.Origins = []*authenticationv1alpha1api.OriginAuthenticationMethod{
					{
						XXX_NoUnkeyedLiteral: struct{}{},
						XXX_unrecognized:     nil,
						XXX_sizecache:        0,
					},
				}
				return p
			},
			true,
		},
		"not equal when labels differ": {
			prep: func() *authenticationv1alpha1.Policy {
				p := policy.DeepCopy()
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
			testPol := tc.prep()
			if policyEqual(testPol, policy) != tc.expect {
				t.Errorf("Expected output to be %t", tc.expect)
			}
		})
	}
}
