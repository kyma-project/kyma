package authz

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPrepareAttributes(t *testing.T) {

	Convey("When no arg is required", t, func() {

		gqlAttributes := noArgsAttributes()
		authAttributes := PrepareAttributes(noArgsContext(), &userInfo, gqlAttributes)

		verifyCommonAttributes(authAttributes)

		Convey("Namespace is empty", func() {
			So(authAttributes.GetNamespace(), ShouldBeEmpty)
		})
		Convey("Name is empty", func() {
			So(authAttributes.GetName(), ShouldBeEmpty)
		})

	})

	Convey("When args are required", t, func() {
		gqlAttributes := withArgsAttributes()
		authAttributes := PrepareAttributes(withArgsContext(), &userInfo, gqlAttributes)

		verifyCommonAttributes(authAttributes)

		Convey("Namespace is set", func() {
			So(authAttributes.GetNamespace(), ShouldEqual, namespace)
		})
		Convey("Name is set", func() {
			So(authAttributes.GetName(), ShouldEqual, name)
		})
	})
}
