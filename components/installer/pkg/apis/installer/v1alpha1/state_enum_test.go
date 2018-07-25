package v1alpha1

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestStatusDeserialization(t *testing.T) {
	Convey("InstallationStatus", t, func() {
		Convey("should deserialize correctly from JSON format", func() {
			var testStatus = InstallationStatus{
				State:       StateInstalled,
				Description: "abcd",
			}

			bytes, err := json.Marshal(testStatus)
			So(err, ShouldBeNil)

			//when
			var s = InstallationStatus{}
			err = json.Unmarshal(bytes, &s)

			//then
			So(err, ShouldBeNil)
			So(s.State, ShouldNotBeNil)
			So(s.State, ShouldEqual, StateInstalled)
			So(s.Description, ShouldEqual, "abcd")
		})
	})
}
