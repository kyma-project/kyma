package kymahelm_test

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymahelm"
	. "github.com/smartystreets/goconvey/convey"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
)

const (
	TestProfileName    = "test"
	ChartDefaultValues = `
existingField: "old_value"
someVar:
  subVar1: "val1"
  subVar2: "val2"
overridedField: "override_me"
`
	ProfileValues = `
newField: "new_value"
overridedField: "foobar"
`
)

func TestGetProfileValues(t *testing.T) {

	Convey("GetProfileValues", t, func() {

		Convey("should activate profile, if profile is found", func() {
			chart, err := getChart()
			So(err, ShouldBeNil)
			So(chart, ShouldNotBeNil)
			So(chart, ShouldNotBeEmpty)

			comboValues, err := kymahelm.GetProfileValues(chart, TestProfileName)
			So(err, ShouldBeNil)

			So(comboValues, ShouldNotBeEmpty)
			So(len(comboValues), ShouldEqual, 2)
			So(comboValues["newField"], ShouldEqual, "new_value")
			So(comboValues["overridedField"], ShouldEqual, "foobar")
		})

		Convey("should use default values, if profile is not found", func() {
			chart, err := getChart()
			So(err, ShouldBeNil)
			So(chart, ShouldNotBeNil)
			So(chart, ShouldNotBeEmpty)

			comboValues, err := kymahelm.GetProfileValues(chart, "not-existing-profile")
			So(err, ShouldBeNil)

			So(comboValues, ShouldNotBeEmpty)
			So(len(comboValues), ShouldEqual, 3)
			So(comboValues["existingField"], ShouldEqual, "old_value")
			So(comboValues["overridedField"], ShouldEqual, "override_me")
		})
	})
}

func getChart() (chart.Chart, error) {
	profileFile := chart.File{
		Name: fmt.Sprintf("profile-%s.yaml", TestProfileName),
		Data: []byte(ProfileValues),
	}

	values, err := chartutil.ReadValues([]byte(ChartDefaultValues))
	if err != nil {
		return chart.Chart{}, err
	}
	return chart.Chart{
		Values: values,
		Files: []*chart.File{
			&profileFile,
		},
	}, nil
}
