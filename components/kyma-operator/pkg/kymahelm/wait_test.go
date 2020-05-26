package kymahelm

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	rls "k8s.io/helm/pkg/proto/hapi/services"
)

func TestWaitForCondition(t *testing.T) {

	Convey("Client.WaitForCondition function should", t, func() {

		rsf := func(releaseName string) (*rls.GetReleaseStatusResponse, error) {
			if releaseName == "some" {
				return nil, nil
			}
			return nil, errors.New("Unknown revision")
		}

		Convey("return a status on first iteration", func() {
			//given
			c := &Client{}
			pf := func(*rls.GetReleaseStatusResponse, error) (bool, error) {
				//always succeeds
				return true, nil
			}

			//when
			b, err := c.WaitForCondition("some", pf, ReleaseStatusFunc(rsf))

			//then
			So(err, ShouldBeNil)
			So(b, ShouldBeTrue)
		})

		Convey("return a status on second try", func() {
			count := 1
			//given
			c := &Client{}
			pf := func(*rls.GetReleaseStatusResponse, error) (bool, error) {
				if count == 2 {
					return true, nil
				}
				count++
				return false, nil
			}

			//when
			b, err := c.WaitForCondition("some", pf, ReleaseStatusFunc(rsf), SleepTimeSecs(0))

			//then
			So(err, ShouldBeNil)
			So(b, ShouldBeTrue)
			So(count, ShouldEqual, 2)
		})

		Convey("return a status on fourth try", func() {
			count := 1
			//given
			c := &Client{}
			pf := func(*rls.GetReleaseStatusResponse, error) (bool, error) {
				if count == 4 {
					return true, nil
				}
				count++
				return false, nil
			}

			//when
			b, err := c.WaitForCondition("some", pf, ReleaseStatusFunc(rsf), SleepTimeSecs(0))

			//then
			So(err, ShouldBeNil)
			So(b, ShouldBeTrue)
			So(count, ShouldEqual, 4)
		})

		Convey("eventually give up trying", func() {
			count := 1
			//given
			c := &Client{}
			pf := func(*rls.GetReleaseStatusResponse, error) (bool, error) {
				count++
				return false, nil
			}

			//when
			b, err := c.WaitForCondition("some", pf, ReleaseStatusFunc(rsf), SleepTimeSecs(0), MaxIterations(3))

			//then
			So(err, ShouldBeNil)
			So(b, ShouldBeFalse)
			So(count, ShouldEqual, 4)
		})

		Convey("return an error as soon as predicate func returns error", func() {
			count := 1
			//given
			c := &Client{}
			pf := func(*rls.GetReleaseStatusResponse, error) (bool, error) {
				if count == 4 {
					return false, errors.New("Predicate error occured")
				}
				count++
				return false, nil
			}

			//when
			b, err := c.WaitForCondition("some", pf, ReleaseStatusFunc(rsf), SleepTimeSecs(0), MaxIterations(10))

			//then
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Predicate error occured")
			So(b, ShouldBeFalse)
			So(count, ShouldEqual, 4)
		})

		Convey("pass an error from the release status func into predicate func", func() {
			count := 1
			//given
			c := &Client{}
			pf := func(resp *rls.GetReleaseStatusResponse, err error) (bool, error) {
				if err != nil {
					So(err.Error(), ShouldContainSubstring, "Release status function error occured")
					return true, nil
				}
				return false, nil
			}

			rsfErr := func(releaseName string) (*rls.GetReleaseStatusResponse, error) {
				if count == 4 {
					return nil, errors.New("Release status function error occured")
				}
				count++

				return nil, nil
			}
			//when
			b, err := c.WaitForCondition("some", pf, ReleaseStatusFunc(rsfErr), SleepTimeSecs(0), MaxIterations(10))

			//then
			So(err, ShouldBeNil)
			So(b, ShouldBeTrue)
			So(count, ShouldEqual, 4)
		})
	})
}
