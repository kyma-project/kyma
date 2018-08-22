package config

import (
	"testing"

	"github.com/kyma-project/kyma/components/installer/pkg/apis/installer/v1alpha1"
	. "github.com/smartystreets/goconvey/convey"
)

func TestConvertTopMap(t *testing.T) {

	Convey("The function consumes a list of KymaComponent structs and converts it to a string-struct map", t, func() {

		Convey("Should the list be empty, the map must not contain any keys", func() {

			fakeComponentList := []v1alpha1.KymaComponent{}
			testMap := convertToMap(fakeComponentList)

			So(len(testMap), ShouldEqual, 0)
		})

		Convey("Should the list contain just one element, the map must be of length 1 as well", func() {

			fakeComponentList := []v1alpha1.KymaComponent{
				v1alpha1.KymaComponent{Name: "core"},
			}
			testMap := convertToMap(fakeComponentList)
			_, exists := testMap["core"]

			So(exists, ShouldBeTrue)
			So(len(testMap), ShouldEqual, 1)
		})

		Convey("Length of the output map should always reflect the number of valid elements", func() {

			fakeComponentList := []v1alpha1.KymaComponent{
				v1alpha1.KymaComponent{Name: "istio"},
				v1alpha1.KymaComponent{Name: "dex"},
				v1alpha1.KymaComponent{Name: "core"},
			}

			testMap := convertToMap(fakeComponentList)

			So(len(testMap), ShouldEqual, 3)
		})
	})
}
