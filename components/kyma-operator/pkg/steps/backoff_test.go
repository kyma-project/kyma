package steps

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBackoff(t *testing.T) {

	Convey("backoffController", t, func() {

		Convey("newBackOff function", func() {

			Convey("should fail with nil intervals", func() {
				_, err := newBackOff(nil, nil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldResemble, "Not enough intervals")
			})

			Convey("should fail with empty intervals", func() {
				_, err := newBackOff([]uint{}, nil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldResemble, "Not enough intervals")
			})

			Convey("should fail without onStep function", func() {
				_, err := newBackOff([]uint{1}, nil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldResemble, "onStep function is missing")
			})

			Convey("should succeed with intervals", func() {
				cnt := 0
				onStepFunc := func(count, max, delay int, msg ...string) {
					cnt++
				}
				res, err := newBackOff([]uint{0}, onStepFunc)
				So(err, ShouldBeNil)

				So(cnt, ShouldEqual, 0)

				res.step()
				So(cnt, ShouldEqual, 1)
				res.step()
				So(cnt, ShouldEqual, 2)
			})
		})

		Convey("onStep function", func() {
			Convey("should be invoked with current and max iteration index", func() {

				var stepRec, maxRec, delayRec int

				onStepFunc := func(step, max, delay int, msg ...string) {
					stepRec = step
					maxRec = max
					delayRec = delay
				}

				res, err := newBackOff([]uint{2, 3, 4, 9}, onStepFunc)
				So(err, ShouldBeNil)
				res.sleepFunc = nil //No delays in test

				So(stepRec, ShouldEqual, 0)
				So(maxRec, ShouldEqual, 0)
				So(delayRec, ShouldEqual, 0)

				res.step()
				So(stepRec, ShouldEqual, 0)  //First step has value of zero
				So(maxRec, ShouldEqual, 3)   //len(intervals) -1
				So(delayRec, ShouldEqual, 2) //First configured delay

				res.step()
				So(stepRec, ShouldEqual, 1)
				So(maxRec, ShouldEqual, 3)
				So(delayRec, ShouldEqual, 3)

				res.step()
				So(stepRec, ShouldEqual, 2)
				So(maxRec, ShouldEqual, 3)
				So(delayRec, ShouldEqual, 4)

				res.step()
				So(stepRec, ShouldEqual, 3) //Step value is now equal to the number of configured delays (array index)
				So(maxRec, ShouldEqual, 3)
				So(delayRec, ShouldEqual, 9) //Last configured delay

				res.step()
				So(stepRec, ShouldEqual, 4) //step value is now greater than number of configured delays (array index)
				So(maxRec, ShouldEqual, 3)
				So(delayRec, ShouldEqual, 9) //Last configured delay
			})
		})
	})
}
