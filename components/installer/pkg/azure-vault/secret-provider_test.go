package azurevault

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRemoveNewLines(t *testing.T) {

	Convey("Values fetched from vault", t, func() {

		testValue := `
first
second
third
`
		Convey("should be converted to one-line string", func() {

			processedValue := removeNewLines(testValue)
			numOfNewlines := strings.Count(processedValue, "\n")

			expectedOutput := "firstsecondthird"

			So(numOfNewlines, ShouldEqual, 0)
			So(processedValue, ShouldEqual, expectedOutput)
		})
	})
}

func TestUseString(t *testing.T) {

	Convey("Pointer to a string", t, func() {

		var pointerToTestString *string

		Convey("should be dereferenced to empty string if it is nil", func() {

			pointerToTestString = nil

			emptyString := ""
			dereferencedPointer := useString(pointerToTestString)

			So(dereferencedPointer, ShouldEqual, emptyString)
		})

		Convey("should be dereferenced to a string if it is not nil", func() {

			testString := "test string"
			pointerToTestString = &testString

			dereferencedPointer := useString(pointerToTestString)

			So(dereferencedPointer, ShouldEqual, testString)
		})
	})
}
