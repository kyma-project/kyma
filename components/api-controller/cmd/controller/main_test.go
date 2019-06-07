package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateApi(t *testing.T) {

	Convey("If readBlacklistedServices", t, func() {

		Convey("is fed with an empty string", func() {
			//given
			list := ""

			Convey("it should return an empty array", func() {

				//when
				blacklistedSvc := readBlacklistedServices(list)

				//then
				So(blacklistedSvc, ShouldBeEmpty)
				So(len(blacklistedSvc), ShouldEqual, 0)
			})
		})

		Convey("is fed with string containing service", func() {
			//given
			list := "svc-1.ns-1"

			Convey("it should return an array containing one item", func() {

				//when
				blacklistedSvc := readBlacklistedServices(list)

				//then
				So(blacklistedSvc, ShouldNotBeEmpty)
				So(len(blacklistedSvc), ShouldEqual, 1)
			})
		})

		Convey("is fed with string containing services", func() {
			//given
			list := "svc-1.ns-1, svc-2.ns-1"

			Convey("it should return an array containing two items", func() {

				//when
				blacklistedSvc := readBlacklistedServices(list)

				//then
				So(blacklistedSvc, ShouldNotBeEmpty)
				So(len(blacklistedSvc), ShouldEqual, 2)
				So(blacklistedSvc[0], ShouldEqual, "svc-1.ns-1")
				So(blacklistedSvc[1], ShouldEqual, "svc-2.ns-1")
			})
		})

		Convey("is fed with string containing an empty item", func() {
			//given
			list := "svc-1.ns-1 , , svc-2.ns-1"

			Convey("it should return an array containing two items", func() {

				//when
				blacklistedSvc := readBlacklistedServices(list)

				//then
				So(blacklistedSvc, ShouldNotBeEmpty)
				So(len(blacklistedSvc), ShouldEqual, 2)
				So(blacklistedSvc[0], ShouldEqual, "svc-1.ns-1")
				So(blacklistedSvc[1], ShouldEqual, "svc-2.ns-1")
			})
		})
	})
}

