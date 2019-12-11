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
	"testing"

	"github.com/google/go-cmp/cmp"

	"knative.dev/pkg/ptr"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	tNs   = "testns"
	tName = "testname"
)

var tOwner = &metav1.OwnerReference{
	APIVersion:         "fakegroup.fakeapi/v1",
	Kind:               "FakeKind",
	Name:               "fake",
	UID:                "00000000-0000-0000-0000-000000000000",
	Controller:         ptr.Bool(true),
	BlockOwnerDeletion: ptr.Bool(true),
}

func TestMetaObjectOptions(t *testing.T) {
	t.Logf("%+s", tOwner)

	objMeta := NewChannel(tNs, tName,
		WithLabel("test.label/2", "val2"),
		WithControllerRef(tOwner),
		WithLabel("test.label/1", "val1"),
	).ObjectMeta

	expectObjMeta := metav1.ObjectMeta{
		Namespace:       tNs,
		Name:            tName,
		OwnerReferences: []metav1.OwnerReference{*tOwner},
		Labels: map[string]string{
			"test.label/1": "val1",
			"test.label/2": "val2",
		},
	}

	if d := cmp.Diff(expectObjMeta, objMeta); d != "" {
		t.Errorf("Unexpected diff: (-:expect, +:got) %s", d)
	}
}
