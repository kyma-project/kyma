package kymahelm

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestWaitForCondition(t *testing.T) {

	const testReleaseName = "someRelease"
	const testReleaseNamespace = "someNamespace"
	testRelNmspcName := NamespacedName{
		Name:      testReleaseName,
		Namespace: testReleaseNamespace,
	}

	mockReleaseStatusFn := ReleaseStatusFunc(func(nn NamespacedName) (*ReleaseStatus, error) {
		if nn.Name == testReleaseName {
			return nil, nil
		}
		return nil, errors.New("Unknown release")
	})

	Convey("Client.WaitForCondition function should", t, func() {

		commonOptions := []WaitOption{mockReleaseStatusFn, SleepTimeSecs(0), MaxIterations(5)}

		Convey("succeed on the first iteration in a happy path scenario", func() {
			//given
			count := 0
			c := &Client{}
			predicateFn := func(*ReleaseStatus, error) (bool, error) {
				count++
				//always succeeds
				return true, nil
			}

			//when
			success, err := c.WaitForCondition(testRelNmspcName, predicateFn, commonOptions...)

			//then
			So(err, ShouldBeNil)
			So(success, ShouldBeTrue)
			So(count, ShouldEqual, 1)
		})

		Convey("succeed on the second try (one retry)", func() {
			//given
			const initialCountValue = 1
			const succeedAtCount = 2
			count := initialCountValue
			c := &Client{}
			predicateFn := succeedAtCountPredicateFn(&count, succeedAtCount)

			//when
			success, err := c.WaitForCondition(testRelNmspcName, predicateFn, commonOptions...)

			//then
			So(err, ShouldBeNil)
			So(success, ShouldBeTrue)
			So(count, ShouldEqual, succeedAtCount)
		})

		Convey("succeed on the fourth try (three retries)", func() {

			//given
			const initialCountValue = 1
			const succeedAtCount = 4
			count := initialCountValue
			c := &Client{}
			predicateFn := succeedAtCountPredicateFn(&count, succeedAtCount)

			//when
			success, err := c.WaitForCondition(testRelNmspcName, predicateFn, commonOptions...)

			//then
			So(err, ShouldBeNil)
			So(success, ShouldBeTrue)
			So(count, ShouldEqual, succeedAtCount)
		})

		Convey("fail after max number of iterations is reached", func() {
			//given
			const initialCountValue = 1
			const expectedRetries = 5
			count := initialCountValue
			c := &Client{}
			predicateFn := alwaysFailPredicateFn(&count)

			//when
			b, err := c.WaitForCondition(testRelNmspcName, predicateFn, commonOptions...)

			//then
			So(err, ShouldBeNil)
			So(b, ShouldBeFalse)
			So(count, ShouldEqual, initialCountValue+expectedRetries)
		})

		Convey("return an error as soon as predicate func returns error", func() {
			//given
			count := 1
			const failAtCount = 4

			c := &Client{}
			failAtCountPredicateFn := func(*ReleaseStatus, error) (bool, error) {
				if count == failAtCount {
					return false, errors.New("Predicate error occured")
				}
				count++
				return false, nil
			}

			//when
			b, err := c.WaitForCondition(testRelNmspcName, failAtCountPredicateFn, commonOptions...)

			//then
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Predicate error occured")
			So(b, ShouldBeFalse)
			So(count, ShouldEqual, failAtCount)
		})

		Convey("pass an error from the release status func into predicate func", func() {
			//given
			const initialCountValue = 1
			const expectedRetries = 4
			count := initialCountValue
			c := &Client{}

			failUntilErrorPredicateFn := func(relStatus *ReleaseStatus, err error) (bool, error) {
				count++
				if err != nil {
					So(err.Error(), ShouldContainSubstring, "Release status function error occured")
					return true, nil
				}
				return false, nil
			}

			relStatusWithErrAtCountFn := func(nn NamespacedName) (*ReleaseStatus, error) {
				if count == expectedRetries {
					return nil, errors.New("Release status function error occured")
				}
				return &ReleaseStatus{}, nil
			}

			//when
			b, err := c.WaitForCondition(testRelNmspcName, failUntilErrorPredicateFn, ReleaseStatusFunc(relStatusWithErrAtCountFn), SleepTimeSecs(0), MaxIterations(10))

			//then
			So(err, ShouldBeNil)
			So(b, ShouldBeTrue)
			So(count, ShouldEqual, initialCountValue+expectedRetries)
		})
	})

	Convey("Client.WaitForCondition function with default number of iterations should", t, func() {

		Convey("fail after max number of iterations is reached", func() {
			//given
			const initialCountValue = 1
			const expectedCountValue = initialCountValue + 10
			count := initialCountValue
			c := &Client{}

			//when
			b, err := c.WaitForCondition(testRelNmspcName, alwaysFailPredicateFn(&count), SleepTimeSecs(0), mockReleaseStatusFn)

			//then
			So(err, ShouldBeNil)
			So(b, ShouldBeFalse)
			So(count, ShouldEqual, expectedCountValue)
		})

	})
}

func alwaysFailPredicateFn(count *int) func(*ReleaseStatus, error) (bool, error) {
	return func(*ReleaseStatus, error) (bool, error) {
		(*count)++
		return false, nil
	}
}

func succeedAtCountPredicateFn(count *int, atValue int) func(*ReleaseStatus, error) (bool, error) {
	return func(*ReleaseStatus, error) (bool, error) {
		if (*count) == atValue {
			return true, nil
		}
		(*count)++
		return false, nil
	}
}
