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

package utils_test

import (
	"testing"

	"github.com/onsi/gomega"

	"github.com/kyma-project/kyma/components/function-controller/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewRuntimeInfo(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dummy",
		},
		Data: map[string]string{
			"dockerRegistry":     "foo",
			"serviceAccountName": "test",
			"runtimes":           "[]",
			"defaults":           "{}",
			"funcTypes":          "[]",
			"funcSizes":          "[]",
		},
	}
	ri, err := utils.New(cm)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(ri.ServiceAccount).To(gomega.Equal("test"))
	g.Expect(ri.RegistryInfo).To(gomega.Equal("foo"))

	cmMissing := &corev1.ConfigMap{
		Data: map[string]string{
			"serviceAccountName": "test",
		},
	}
	_, err = utils.New(cmMissing)
	g.Expect(err).ToNot(gomega.BeNil())
	g.Expect(err.Error()).To(gomega.ContainSubstring("missing mandatory attributes in ConfigMap data"))

	cmBroken := &corev1.ConfigMap{
		Data: map[string]string{
			"serviceAccountName": "test",
			"dockerRegistry":     "foo",
			"runtimes":           "invalid",
			"defaults":           "{}",
			"funcTypes":          "[]",
			"funcSizes":          "[]",
		},
	}
	_, err = utils.New(cmBroken)
	g.Expect(err).ToNot(gomega.BeNil())
	g.Expect(err.Error()).To(gomega.ContainSubstring("error unmarshaling JSON"))
}
