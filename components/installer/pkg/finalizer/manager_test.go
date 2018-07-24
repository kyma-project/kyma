package finalizer

import (
	"testing"

	"github.com/kyma-project/kyma/components/installer/pkg/apis/installer/v1alpha1"
	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

const finalizerName = "test.finalizer.kyma.cx"

func TestAddFinalizer(t *testing.T) {

	Convey("AddFinalizer requires a Kubernetes object as parameter", t, func() {

		var inst *v1alpha1.Installation
		m := NewManager(finalizerName)

		Convey("If the object has no finalizers, the function should update it by adding one", func() {

			inst = createInstallation()

			m.AddFinalizer(inst)

			So(len(inst.Finalizers), ShouldEqual, 1)
			So(finalizerName, ShouldBeIn, inst.Finalizers)
		})

		Convey("If the object has any finalizers, the function should not override them", func() {

			inst = createInstallation("first.finalizer.kyma.cx")

			m.AddFinalizer(inst)

			So(len(inst.Finalizers), ShouldEqual, 2)
			So(finalizerName, ShouldBeIn, inst.Finalizers)
		})
	})
}

func TestRemoveFinalizer(t *testing.T) {

	Convey("RemoveFinalizer requires a Kubernetes object as parameter", t, func() {

		var inst *v1alpha1.Installation
		m := NewManager(finalizerName)

		Convey("If the object has no finalizers at all, it should not be modified", func() {

			inst = createInstallation()

			m.RemoveFinalizer(inst)

			So(len(inst.Finalizers), ShouldEqual, 0)
			So(finalizerName, ShouldNotBeIn, inst.Finalizers)
		})

		Convey("If the object does not have the finalizer specified in Manager, it should not be modified as well", func() {

			inst = createInstallation("random.finalizer.kyma.cx", "another.finalizer.kyma.cx")

			m.RemoveFinalizer(inst)

			So(len(inst.Finalizers), ShouldEqual, 2)
			So(finalizerName, ShouldNotBeIn, inst.Finalizers)
		})

		Convey("If the object has the finalizer specified in Manager, it should be removed", func() {

			inst = createInstallation(finalizerName, "random.kyma.cx")

			m.RemoveFinalizer(inst)

			So(len(inst.Finalizers), ShouldEqual, 1)
			So(finalizerName, ShouldNotBeIn, inst.Finalizers)
		})
	})
}

func createInstallation(finalizers ...string) *v1alpha1.Installation {

	return &v1alpha1.Installation{
		ObjectMeta: v1.ObjectMeta{
			Finalizers: finalizers,
		},
	}
}
