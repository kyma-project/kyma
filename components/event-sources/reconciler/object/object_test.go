package object

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/ptr"
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
