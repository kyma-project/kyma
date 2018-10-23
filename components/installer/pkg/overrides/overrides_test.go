package overrides

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestOverrides(t *testing.T) {

	Convey("joinOverridesMap function", t, func() {

		Convey("Should work with nil input", func() {

			var testInput []inputMap = nil
			actual := joinOverridesMap(testInput...)

			So(actual, ShouldNotBeNil)
			So(len(actual), ShouldEqual, 0)
		})

		Convey("Should work with empty input", func() {

			actual := joinOverridesMap()

			So(actual, ShouldNotBeNil)
			So(len(actual), ShouldEqual, 0)
		})

		Convey("Should work fine with single input", func() {
			m1 := makeTestMap("key1:val1_1", "key2:val1_2", "key3:val1_3")

			actual := joinOverridesMap(m1)

			So(actual, ShouldNotBeNil)
			So(len(actual), ShouldEqual, 3)
			So(actual["key1"], ShouldEqual, "val1_1")
			So(actual["key2"], ShouldEqual, "val1_2")
			So(actual["key3"], ShouldEqual, "val1_3")
		})

		Convey("Should join two maps without overlaps", func() {
			m1 := makeTestMap("key1:val1_1", "key2:val1_2", "key3:val1_3")
			m2 := makeTestMap("key4:val2_1", "key5:val2_2", "key6:val2_3")

			actual := joinOverridesMap(m1, m2)

			So(actual, ShouldNotBeNil)
			So(len(actual), ShouldEqual, 6)
			So(actual["key1"], ShouldEqual, "val1_1")
			So(actual["key2"], ShouldEqual, "val1_2")
			So(actual["key3"], ShouldEqual, "val1_3")
			So(actual["key4"], ShouldEqual, "val2_1")
			So(actual["key5"], ShouldEqual, "val2_2")
			So(actual["key6"], ShouldEqual, "val2_3")
		})

		Convey("Should join three maps without overlaps", func() {
			m1 := makeTestMap("key1:val1_1", "key2:val1_2", "key3:val1_3")
			m2 := makeTestMap("key4:val2_1", "key5:val2_2", "key6:val2_3")
			m3 := makeTestMap("key7:val3_1", "key8:val3_2", "key9:val3_3")

			actual := joinOverridesMap([]inputMap{m1, m2, m3}...)

			So(actual, ShouldNotBeNil)
			So(len(actual), ShouldEqual, 9)
			So(actual["key1"], ShouldEqual, "val1_1")
			So(actual["key2"], ShouldEqual, "val1_2")
			So(actual["key3"], ShouldEqual, "val1_3")
			So(actual["key4"], ShouldEqual, "val2_1")
			So(actual["key5"], ShouldEqual, "val2_2")
			So(actual["key6"], ShouldEqual, "val2_3")
			So(actual["key7"], ShouldEqual, "val3_1")
			So(actual["key8"], ShouldEqual, "val3_2")
			So(actual["key9"], ShouldEqual, "val3_3")
		})

		Convey("Should join two maps with overlaps", func() {
			m1 := makeTestMap("key1:val1_1", "key2:val1_2", "key3:val1_3")
			m2 := makeTestMap("key2:val2_1", "key3:val2_2", "key4:val2_3")

			actual := joinOverridesMap(m1, m2)

			So(actual, ShouldNotBeNil)
			So(len(actual), ShouldEqual, 4)
			So(actual["key1"], ShouldEqual, "val1_1") //from m1
			So(actual["key2"], ShouldEqual, "val2_1") //from m2, last wins
			So(actual["key3"], ShouldEqual, "val2_2") //from m2, last wins
			So(actual["key4"], ShouldEqual, "val2_3") //from m2
		})

		Convey("Should join three maps with overlaps", func() {
			m1 := makeTestMap("key1:val1", "keyY:val_Y1", "keyX:val_X1")
			m2 := makeTestMap("keyX:val_X2", "key5:val5", "keyY:val_Y2")
			m3 := makeTestMap("keyY:val_Y3", "keyX:val_X3", "key9:val9")

			actual := joinOverridesMap(m1, m2, m3)

			So(actual, ShouldNotBeNil)
			So(len(actual), ShouldEqual, 5)
			So(actual["key1"], ShouldEqual, "val1")   //from m1
			So(actual["key5"], ShouldEqual, "val5")   //from m2
			So(actual["key9"], ShouldEqual, "val9")   //from m3
			So(actual["keyX"], ShouldEqual, "val_X3") //from m3, last wins
			So(actual["keyY"], ShouldEqual, "val_Y3") //from m3, last wins
		})
	})

	Convey("joinComponentOverrides function", t, func() {

		Convey("Should work with nil input", func() {

			var testInput []component = nil
			actual := joinComponentOverrides(testInput...)

			So(actual, ShouldNotBeNil)
			So(len(actual), ShouldEqual, 0)
		})

		Convey("Should work with empty input", func() {

			actual := joinComponentOverrides()

			So(actual, ShouldNotBeNil)
			So(len(actual), ShouldEqual, 0)
		})

		Convey("Should work fine with single input", func() {
			c1 := component{name: "test", overrides: makeTestMap("key1:val1_1", "key2:val1_2", "key3:val1_3")}
			res := joinComponentOverrides(c1)

			So(res, ShouldNotBeNil)
			So(len(res), ShouldEqual, 1)
			testOverrides := res["test"]

			So(len(testOverrides), ShouldEqual, 3)
			So(testOverrides["key1"], ShouldEqual, "val1_1")
			So(testOverrides["key2"], ShouldEqual, "val1_2")
			So(testOverrides["key3"], ShouldEqual, "val1_3")
		})

		Convey("Should handle two different components", func() {
			c1 := component{
				name:      "test1",
				overrides: makeTestMap("key1:val1_1", "key2:val1_2", "key3:val1_3"),
			}
			c2 := component{
				name:      "test2",
				overrides: makeTestMap("key4:val2_1", "key5:val2_2", "key6:val2_3"),
			}

			res := joinComponentOverrides(c1, c2)

			So(res, ShouldNotBeNil)
			So(len(res), ShouldEqual, 2)
			test1Overrides := res["test1"]

			So(len(test1Overrides), ShouldEqual, 3)
			So(test1Overrides["key1"], ShouldEqual, "val1_1")
			So(test1Overrides["key2"], ShouldEqual, "val1_2")
			So(test1Overrides["key3"], ShouldEqual, "val1_3")

			test2Overrides := res["test2"]
			So(len(test2Overrides), ShouldEqual, 3)
			So(test2Overrides["key4"], ShouldEqual, "val2_1")
			So(test2Overrides["key5"], ShouldEqual, "val2_2")
			So(test2Overrides["key6"], ShouldEqual, "val2_3")
		})

		Convey("Should join two inputs for the same component", func() {

			c1 := component{
				name:      "test",
				overrides: makeTestMap("key1:val1_1", "key2:val1_2", "key3:val1_3"),
			}
			c2 := component{
				name:      "test",
				overrides: makeTestMap("key3:val2_1", "key4:val2_2", "key5:val2_3"),
			}

			res := joinComponentOverrides(c1, c2)

			So(res, ShouldNotBeNil)
			So(len(res), ShouldEqual, 1)
			testOverrides := res["test"]

			So(len(testOverrides), ShouldEqual, 5)
			So(testOverrides["key1"], ShouldEqual, "val1_1") //from c1
			So(testOverrides["key2"], ShouldEqual, "val1_2") //from c1
			So(testOverrides["key3"], ShouldEqual, "val2_1") //from c2, last wins
			So(testOverrides["key4"], ShouldEqual, "val2_2") //from c2
			So(testOverrides["key5"], ShouldEqual, "val2_3") //from c2
		})

		Convey("Should join multiple inputs for multiple components", func() {
			c1 := component{
				name:      "test1",
				overrides: makeTestMap("key1:val1_1", "key2:val1_2", "key3:val1_3"),
			}
			c2 := component{
				name:      "test2",
				overrides: makeTestMap("key1:val2_1", "key2:val2_2", "key3:val2_3"),
			}
			c3 := component{
				name:      "test1",
				overrides: makeTestMap("key3:val3_1", "key4:val3_2", "key5:val3_3"),
			}
			c4 := component{
				name:      "test2",
				overrides: makeTestMap("key3:val4_1", "key4:val4_2", "key5:val4_3"),
			}
			res := joinComponentOverrides(c1, c2, c3, c4)

			So(res, ShouldNotBeNil)
			So(len(res), ShouldEqual, 2)

			test1Overrides := res["test1"]

			So(len(test1Overrides), ShouldEqual, 5)
			So(test1Overrides["key1"], ShouldEqual, "val1_1") //from c1
			So(test1Overrides["key2"], ShouldEqual, "val1_2") //from c1
			So(test1Overrides["key3"], ShouldEqual, "val3_1") //from c3, last wins
			So(test1Overrides["key4"], ShouldEqual, "val3_2") //from c3
			So(test1Overrides["key5"], ShouldEqual, "val3_3") //from c3

			test2Overrides := res["test2"]

			So(len(test2Overrides), ShouldEqual, 5)
			So(test2Overrides["key1"], ShouldEqual, "val2_1") //from c2
			So(test2Overrides["key2"], ShouldEqual, "val2_2") //from c2
			So(test2Overrides["key3"], ShouldEqual, "val4_1") //from c4, last wins
			So(test2Overrides["key4"], ShouldEqual, "val4_2") //from c4
			So(test2Overrides["key5"], ShouldEqual, "val4_3") //from c4
		})
	})
}

func makeTestMap(entries ...string) inputMap {
	res := make(inputMap)

	if entries == nil {
		return res
	}

	for _, s := range entries {
		entry := strings.Split(s, ":")
		if len(entry) != 2 {
			panic("Invalid input - expected two string values separated by a colon, got: " + s)
		}
		res[entry[0]] = entry[1]
	}

	return res
}
