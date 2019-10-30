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

package objects

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

const fixtureKsvcPath = "../../test/fixtures/ksvc.json"

var fixtureKsvc *servingv1.Service

func TestMain(m *testing.M) {
	var err error

	fixtureKsvc, err = loadFixtureKsvc()
	if err != nil {
		panic(errors.Wrap(err, "loading Knative Service from fixtures"))
	}

	os.Exit(m.Run())
}

func loadFixtureKsvc() (*servingv1.Service, error) {
	data, err := ioutil.ReadFile(fixtureKsvcPath)
	if err != nil {
		return nil, err
	}

	ksvc := &servingv1.Service{}
	if err := json.Unmarshal(data, ksvc); err != nil {
		return nil, err
	}

	return ksvc, nil
}

func TestKsvcEqual(t *testing.T) {
	ksvc := fixtureKsvc

	if !ksvcEqual(nil, nil) {
		t.Error("Two nil elements should be equal")
	}

	testCases := map[string]struct {
		prep   func() *servingv1.Service
		expect bool
	}{
		"not equal when one element is nil": {
			func() *servingv1.Service {
				return nil
			},
			false,
		},
		"not equal when labels differ": {
			func() *servingv1.Service {
				ksvcCopy := ksvc.DeepCopy()
				ksvcCopy.Labels["foo"] += "test"
				return ksvcCopy
			},
			false,
		},
		"not equal when annotations differ": {
			func() *servingv1.Service {
				ksvcCopy := ksvc.DeepCopy()
				ksvcCopy.Annotations["foo"] += "test"
				return ksvcCopy
			},
			false,
		},
		"equal when other fields differ": {
			func() *servingv1.Service {
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

				ps := sp.ConfigurationSpec.Template.Spec.PodSpec

				*sp = servingv1.ServiceSpec{} // reset
				sp.ConfigurationSpec.Template.Spec.PodSpec = ps

				// status
				st := &ksvcCopy.Status
				*st = servingv1.ServiceStatus{} // reset

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
		"not equal when envvar differs": {
			func() *corev1.Container {
				cCopy := c.DeepCopy()
				cCopy.Env[0].Value += "test"
				return cCopy
			},
			false,
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

func TestResourceListEqual(t *testing.T) {
	cs := fixtureKsvc.Spec.ConfigurationSpec.Template.Spec.PodSpec.Containers

	if cs == nil {
		t.Fatalf("Test requires fixture object %s to have a least one Container", fixtureKsvcPath)
	}
	if cs[0].Resources.Requests == nil {
		t.Fatalf("Test requires fixture object %s to have a least one Request", fixtureKsvcPath)
	}
	var hasBinaryRequest bool
	for r := range cs[0].Resources.Requests {
		if cs[0].Resources.Requests[r].Format == resource.BinarySI {
			hasBinaryRequest = true
			break
		}
	}
	if !hasBinaryRequest {
		t.Fatalf("Test requires fixture object %s to have a least one Request represented in the binary SI", fixtureKsvcPath)
	}

	rl := fixtureKsvc.Spec.ConfigurationSpec.Template.Spec.PodSpec.Containers[0].Resources.Requests

	testCases := map[string]struct {
		prep   func() corev1.ResourceList
		expect bool
	}{
		"not equal when resource types differ": {
			func() corev1.ResourceList {
				rlCopy := rl.DeepCopy()
				for k, v := range rlCopy {
					delete(rlCopy, k)
					rlCopy[corev1.ResourceName(k+"test")] = v
				}
				return rlCopy
			},
			false,
		},
		"not equal when some resource differs": {
			func() corev1.ResourceList {
				rlCopy := rl.DeepCopy()
				for k, v := range rlCopy {
					(&v).Add(*resource.NewQuantity(1, resource.DecimalSI))
					rlCopy[k] = v
				}
				return rlCopy
			},
			false,
		},
		"equal when resources are semantically equal": {
			func() corev1.ResourceList {
				rlCopy := rl.DeepCopy()
				for k, v := range rlCopy {
					var newVal *resource.Quantity
					switch v.Format {
					case resource.DecimalSI:
						newVal = resource.NewMilliQuantity(v.MilliValue(), resource.BinarySI)
					default:
						newVal = resource.NewMilliQuantity(v.MilliValue(), resource.DecimalSI)
					}
					rlCopy[k] = *newVal
				}
				return rlCopy
			},
			true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			testRl := tc.prep()
			if resourceListEqual(testRl, rl) != tc.expect {
				t.Errorf("Expected output to be %t", tc.expect)
			}
		})
	}
}
