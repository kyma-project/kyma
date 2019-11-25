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
	"testing"

	"github.com/google/go-cmp/cmp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"

	sourcesv1alpha1 "github.com/kyma-project/kyma/components/event-sources/apis/sources/v1alpha1"
)

func TestNewChannel(t *testing.T) {
	const (
		ns   = "testns"
		name = "test"
	)

	testHTTPSrc := &sourcesv1alpha1.HTTPSource{ObjectMeta: metav1.ObjectMeta{
		Namespace: ns,
		Name:      name,
		UID:       "00000000-0000-0000-0000-000000000000",
	}}

	expectCh := &messagingv1alpha1.Channel{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
			OwnerReferences: []metav1.OwnerReference{
				*testHTTPSrc.ToOwner(),
			},
		},
	}

	ch := NewChannel(ns, name,
		WithChannelControllerRef(testHTTPSrc.ToOwner()),
	)

	if d := cmp.Diff(expectCh, ch); d != "" {
		t.Errorf("Unexpected diff: (-:expect, +:got) %s", d)
	}
}
