package config

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConvertTopMap(t *testing.T) {

	Convey("The function consumes a string of comma-separated values and converts it to a string-struct map", t, func() {

		var fakeComponentList string

		Convey("Should the string be empty, the map must not contain any keys", func() {

			fakeComponentList = ""
			testMap := convertToMap(fakeComponentList)

			So(len(testMap), ShouldEqual, 0)
		})

		Convey("Should the string contain just one element, the map must be of length 1 as well", func() {

			fakeComponentList = "core"
			testMap := convertToMap(fakeComponentList)
			_, exists := testMap["core"]

			So(exists, ShouldBeTrue)
			So(len(testMap), ShouldEqual, 1)
		})

		Convey("Length of the output map should always reflect the number of valid elements", func() {

			fakeComponentList = "istio,dex,core"
			testMap := convertToMap(fakeComponentList)

			So(len(testMap), ShouldEqual, 3)
		})

		Convey("In case of a trailing comma, the map should not create additional key", func() {

			fakeComponentList = "dex,core,"
			testMap := convertToMap(fakeComponentList)
			_, exists := testMap[""]

			So(exists, ShouldBeFalse)
			So(len(testMap), ShouldEqual, 2)
		})

		Convey("In case of a leading comma, the map should not create additional key as well", func() {

			fakeComponentList = ",dex,core"
			testMap := convertToMap(fakeComponentList)
			_, exists := testMap[""]

			So(exists, ShouldBeFalse)
			So(len(testMap), ShouldEqual, 2)
		})

		Convey("Any spaces should be removed before processing the string", func() {

			fakeComponentList = "dex, ,core,   ,"
			testMap := convertToMap(fakeComponentList)
			_, oneSpaceKeyExists := testMap[" "]
			_, threeSpacesKeyExists := testMap["   "]

			So(oneSpaceKeyExists, ShouldBeFalse)
			So(threeSpacesKeyExists, ShouldBeFalse)
			So(len(testMap), ShouldEqual, 2)
		})
	})
}
