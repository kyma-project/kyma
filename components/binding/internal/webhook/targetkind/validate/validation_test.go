/*
Copyright 2017 The Kubernetes Authors.

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

package validate

import (
	"github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func validTargetKind() *v1alpha1.TargetKind {
	return &v1alpha1.TargetKind{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-binding",
			Namespace: "test-ns",
		},
		Spec: v1alpha1.TargetKindSpec{},
	}
}

func TestValidateTargetKind(t *testing.T) {
	cases := []struct {
		name    string
		binding *v1alpha1.TargetKind
		valid   bool
	}{
		{
			name:    "valid",
			binding: validTargetKind(),
			valid:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			errs := internalValidateTargetKind(tc.binding)
			if len(errs) != 0 && tc.valid {
				t.Errorf("unexpected error: %v", errs)
			} else if len(errs) == 0 && !tc.valid {
				t.Error("unexpected success")
			}
		})
	}
}
