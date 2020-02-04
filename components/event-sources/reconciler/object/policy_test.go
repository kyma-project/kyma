/*
Copyright 2020 The Kyma Authors.

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
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/google/go-cmp/cmp"
	authenticationv1alpha1api "istio.io/api/authentication/v1alpha1"
	authenticationv1alpha1 "istio.io/client-go/pkg/apis/authentication/v1alpha1"
)

const (
	tTarget = "tRev-private"
)

func TestNewPolicy(t *testing.T) {
	policy := NewPolicy(tNs, tName,
		WithTarget(tTarget))

	expectPolicy := &authenticationv1alpha1.Policy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: tNs,
			Name:      tName,
		},
		Spec: authenticationv1alpha1api.Policy{
			Targets: []*authenticationv1alpha1api.TargetSelector{
				{Name: tTarget},
			},
		},
	}

	if d := cmp.Diff(expectPolicy, policy); d != "" {
		t.Errorf("Unexpected diff: (-:expect, +:got) %s", d)
	}
}

func TestApplyExistingPolicyAttributes(t *testing.T) {
	existingPolicy := &authenticationv1alpha1.Policy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       tNs,
			Name:            tName,
			ResourceVersion: "100",
		},
		Spec: authenticationv1alpha1api.Policy{
			Targets: []*authenticationv1alpha1api.TargetSelector{
				{Name: tTarget},
			},
		},
	}

	desiredPolicy := NewPolicy(tNs, tName,
		WithTarget(tTarget))

	ApplyExistingPolicyAttributes(existingPolicy, desiredPolicy)
	expectedPolicy := &authenticationv1alpha1.Policy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       tNs,
			Name:            tName,
			ResourceVersion: "100",
		},
		Spec: authenticationv1alpha1api.Policy{
			Targets: []*authenticationv1alpha1api.TargetSelector{
				{Name: tTarget},
			},
		},
	}

	if d := cmp.Diff(desiredPolicy, expectedPolicy); d != "" {
		t.Errorf("Unexpected diff: (-:expect, +:got) %s", d)
	}
}
