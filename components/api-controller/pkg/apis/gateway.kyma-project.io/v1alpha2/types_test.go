package v1alpha2

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDeserializeMatchType(t *testing.T) {

	Convey("Function toMatchExpression()", t, func() {

		Convey("Should fail for a nil map", func() {
			mt, err := toMatchExpression(nil)
			So(mt, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, "Expected exactly 1 entry in MatchExpression, got: 0")
		})

		Convey("Should fail for an empty map", func() {
			emptyMap := map[string]string{}
			mt, err := toMatchExpression(emptyMap)
			So(mt, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, "Expected exactly 1 entry in MatchExpression, got: 0")
		})

		Convey("Should fail for a map with two entries", func() {
			twoMap := map[string]string{"a": "1", "b": "2"}
			mt, err := toMatchExpression(twoMap)
			So(mt, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, "Expected exactly 1 entry in MatchExpression, got: 2")
		})

		Convey("Should fail for a map with an unknown type", func() {
			wrongTypeMap := map[string]string{"wrng": "1"}
			mt, err := toMatchExpression(wrongTypeMap)
			So(mt, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, "Unknown MatchExpression type: \"wrng\"")
		})

		Convey("Should unmarshall a correct map", func() {
			correctMap := map[string]string{"exact": "/do/not/use/in/production"}

			mt, err := toMatchExpression(correctMap)
			So(err, ShouldBeNil)
			So(mt, ShouldNotBeNil)
			So(mt.ExprType, ShouldEqual, ExactMatch)
			So(mt.Value, ShouldEqual, correctMap[string(ExactMatch)])
		})
	})
}
