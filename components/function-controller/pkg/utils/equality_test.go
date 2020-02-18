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

package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	stdlog "log"
	"os"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const fixtureKsvcPath = "../../test/fixtures/ksvc.json"

var fixtureKsvc *servingv1.Service

func TestMain(m *testing.M) {
	var err error

	fixtureKsvc, err = loadFixtureKsvc()
	if err != nil {
		stdlog.Fatal("while reading ksvc.json:", err)
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
		"equal when other fields differ": {
			func() *servingv1.Service {
				ksvcCopy := ksvc.DeepCopy()

				// metadata
				lbls := ksvcCopy.Labels

				m := &ksvcCopy.ObjectMeta
				m.Reset()
				m.Labels = lbls

				// spec
				sp := &ksvcCopy.Spec

				ps := sp.ConfigurationSpec.Template.Spec.PodSpec

				*sp = servingv1.ServiceSpec{} // reset
				sp.ConfigurationSpec.Template.Spec.PodSpec = ps

				// status
				ksvcCopy.Status = servingv1.ServiceStatus{}

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
	spec := &fixtureKsvc.Spec.ConfigurationSpec.Template.Spec.PodSpec

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
				tcopy := spec.DeepCopy()
				tcopy.ServiceAccountName += "foo"
				return tcopy
			},
			false,
		},
		"equal when other fields differ": {
			func() *corev1.PodSpec {
				tcopy := spec.DeepCopy()

				containers := tcopy.Containers
				saName := tcopy.ServiceAccountName

				tcopy.Reset()
				tcopy.Containers = containers
				tcopy.ServiceAccountName = saName

				return tcopy
			},
			true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			testSpec := tc.prep()
			if podSpecEqual(testSpec, spec) != tc.expect {
				t.Errorf("Expected output to be %t", tc.expect)
			}
		})
	}
}

func TestContainerEqual(t *testing.T) {
	cs := fixtureKsvc.Spec.ConfigurationSpec.Template.Spec.PodSpec.Containers
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
	rl := fixtureKsvc.Spec.ConfigurationSpec.Template.Spec.PodSpec.Containers[0].Resources.Requests

	testCases := map[string]struct {
		prep   func() corev1.ResourceList
		expect bool
	}{
		"not equal when resource types differ": {
			func() corev1.ResourceList {
				rlCopy := rl.DeepCopy()
				delete(rlCopy, corev1.ResourceCPU)
				rlCopy[corev1.ResourceStorage] = *resource.NewQuantity(1, resource.DecimalSI)
				return rlCopy
			},
			false,
		},
		"not equal when some resource differs": {
			func() corev1.ResourceList {
				rlCopy := rl.DeepCopy()
				cpuQty := rlCopy[corev1.ResourceCPU]
				cpuQty.Add(*resource.NewQuantity(1, resource.DecimalSI))
				rlCopy[corev1.ResourceCPU] = cpuQty
				return rlCopy
			},
			false,
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
