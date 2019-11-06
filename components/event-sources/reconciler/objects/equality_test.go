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
	"io/ioutil"
	"os"
	"testing"

	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"

	servingv1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"
)

const fixtureKsvcPath = "../../test/fixtures/ksvc.json"

var fixtureKsvc *servingv1alpha1.Service

func TestMain(m *testing.M) {
	var err error

	fixtureKsvc, err = loadFixtureKsvc()
	if err != nil {
		panic(errors.Wrap(err, "loading Knative Service from fixtures"))
	}

	os.Exit(m.Run())
}

func loadFixtureKsvc() (*servingv1alpha1.Service, error) {
	data, err := ioutil.ReadFile(fixtureKsvcPath)
	if err != nil {
		return nil, err
	}

	ksvc := &servingv1alpha1.Service{}
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
		prep   func() *servingv1alpha1.Service
		expect bool
	}{
		"not equal when one element is nil": {
			func() *servingv1alpha1.Service {
				return nil
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
		"equal when other fields differ": {
			func() *servingv1alpha1.Service {
				ksvcCopy := ksvc.DeepCopy()

				// metadata
				anns := ksvcCopy.Annotations

				m := &ksvcCopy.ObjectMeta
				m.Reset()
				m.Annotations = anns

				// spec
				sp := &ksvcCopy.Spec

				ps := sp.ConfigurationSpec.Template.Spec.PodSpec

				*sp = servingv1alpha1.ServiceSpec{} // reset
				sp.ConfigurationSpec.Template = &servingv1alpha1.RevisionTemplateSpec{}
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
		"equal when other fields differ": {
			func() *corev1.Container {
				cCopy := c.DeepCopy()

				img := cCopy.Image
				ports := cCopy.Ports
				env := cCopy.Env

				cCopy.Reset()
				cCopy.Image = img
				cCopy.Ports = ports
				cCopy.Env = env

				return cCopy
			},
			true,
		},
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
