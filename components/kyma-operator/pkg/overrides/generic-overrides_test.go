package overrides

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGenericOverrides(t *testing.T) {

	Convey("GenericOverrides", t, func() {

		Convey("MergeMaps function", func() {

			Convey("Should merge two maps with non-overlapping keys", func() {
				const base = `
a:
  b:
    j: "100"
    k: "200"
    l: "300"
`
				const override = `
p:
  q:
    x1: "1100"
    y1: "2100"
    z1: "3100"
`
				const expected = `
a:
  b:
    j: "100"
    k: "200"
    l: "300"
p:
  q:
    x1: "1100"
    y1: "2100"
    z1: "3100"
`
				baseMap, err := ToMap(base)
				So(err, ShouldBeNil)

				overrideMap, err := ToMap(override)
				So(err, ShouldBeNil)

				MergeMaps(baseMap, overrideMap)
				res, err := ToYaml(baseMap)
				So(err, ShouldBeNil)
				So(res, ShouldEqual, nlnl(expected))
			})

			Convey("Should merge two maps with overlapping keys", func() {
				const base = `
a:
  b:
    j: "100"
    k: "200"
    l: 300
`
				const override = `
a:
  b:
    i: "1100"
    j: 100
    k:
      x1: foo
      y1:
        z1: bar
    l: "300"

`
				const expected = `
a:
  b:
    i: "1100"
    j: 100
    k:
      x1: foo
      y1:
        z1: bar
    l: "300"
`
				baseMap, err := ToMap(base)
				So(err, ShouldBeNil)

				overrideMap, err := ToMap(override)
				So(err, ShouldBeNil)

				testMap := Map{}
				MergeMaps(testMap, baseMap)
				MergeMaps(testMap, overrideMap)
				res, err := ToYaml(testMap)
				So(err, ShouldBeNil)
				So(res, ShouldEqual, nlnl(expected))
			})

			Convey("Should merge non-empty map with an empty one", func() {
				const base = `
a:
  b:
    j: "100"
    k: 200
    l: abc
`
				const expected = `
a:
  b:
    j: "100"
    k: 200
    l: abc
`
				baseMap, err := ToMap(base)
				So(err, ShouldBeNil)

				emptyMap, err := ToMap("")
				So(err, ShouldBeNil)
				So(len(emptyMap), ShouldEqual, 0)

				MergeMaps(baseMap, emptyMap)
				res, err := ToYaml(baseMap)
				So(err, ShouldBeNil)
				So(res, ShouldEqual, nlnl(expected))
			})

			Convey("Should merge empty map with non-empty one", func() {
				const override = `
a:
  b:
    j: "100"
    k: 200
    l: abc
`
				const expected = `
a:
  b:
    j: "100"
    k: 200
    l: abc
`
				baseMap, err := ToMap("")
				So(err, ShouldBeNil)
				So(len(baseMap), ShouldEqual, 0)

				overrideMap, err := ToMap(override)
				So(err, ShouldBeNil)

				MergeMaps(baseMap, overrideMap)
				res, err := ToYaml(baseMap)
				So(err, ShouldBeNil)
				So(res, ShouldEqual, nlnl(expected))
			})
		})

		Convey("ToYaml function", func() {

			Convey("Should not fail for empty map", func() {

				inputMap := map[string]string{}
				res, err := ToYaml(UnflattenToMap(inputMap))
				So(err, ShouldBeNil)
				So(res, ShouldBeBlank)
			})

			Convey("Should merge several entries into one yaml", func() {

				const expected = `
a:
  b:
    c: "100"
    d: "200"
    e: "300"
`
				inputMap := map[string]string{}
				inputMap["a.b.c"] = "100"
				inputMap["a.b.d"] = "200"
				inputMap["a.b.e"] = "300"
				res, err := ToYaml(UnflattenToMap(inputMap))
				So(err, ShouldBeNil)
				So(res, ShouldEqual, nlnl(expected))
			})

			Convey("Should handle multi-line string correctly", func() {

				const expected = `
a:
  b:
    c: "100"
    d: "200"
    e: |
      300
      400
      500
`
				inputMap := map[string]string{}
				inputMap["a.b.c"] = "100"
				inputMap["a.b.d"] = "200"
				inputMap["a.b.e"] = "300\n400\n500\n"

				res, err := ToYaml(UnflattenToMap(inputMap))
				So(err, ShouldBeNil)
				So(res, ShouldEqual, nlnl(expected))
			})

			Convey("Should handle global values", func() {

				const expected = `
a:
  b:
    c: "100"
    d: "200"
    e: "300"
global:
  foo: bar
h:
  o:
    o: xyz
`
				inputMap := map[string]string{}
				inputMap["a.b.c"] = "100"
				inputMap["a.b.d"] = "200"
				inputMap["a.b.e"] = "300"
				inputMap["global.foo"] = "bar"
				inputMap["h.o.o"] = "xyz"

				res, err := ToYaml(UnflattenToMap(inputMap))
				So(err, ShouldBeNil)
				So(res, ShouldEqual, nlnl(expected))
			})

		})

		Convey("ToMap function", func() {
			Convey("Should unmarshall yaml into a map", func() {
				const value = `
a:
  b:
    c: "100"
    d: "200"
    e: "300"
`
				res, err := ToMap(value)
				So(err, ShouldBeNil)

				a, ok := res["a"].(map[string]interface{})
				So(ok, ShouldBeTrue)

				b, ok := a["b"].(map[string]interface{})
				So(ok, ShouldBeTrue)

				c, ok := b["c"].(string)
				So(ok, ShouldBeTrue)
				So(c, ShouldEqual, "100")

				d, ok := b["d"].(string)
				So(ok, ShouldBeTrue)
				So(d, ShouldEqual, "200")

				e, ok := b["e"].(string)
				So(ok, ShouldBeTrue)
				So(e, ShouldEqual, "300")
			})

		})

		Convey("FlattenMap function", func() {

			Convey("Should flatten the map", func() {
				const value = `
a:
  b:
    c: "100"
    d: "200"
    e: "300"
`
				oMap, err := ToMap(value)
				So(err, ShouldBeNil)

				res := FlattenMap(oMap)
				So(len(res), ShouldEqual, 3)
				So(res["a.b.c"], ShouldEqual, "100")
				So(res["a.b.d"], ShouldEqual, "200")
				So(res["a.b.e"], ShouldEqual, "300")
			})
		})

		Convey("copyMap function", func() {

			Convey("Should copy nested map(s)", func() {
				const value = `
a:
  b:
    c: 100
    d:
      e: "200"
`
				src, err := ToMap(value)
				So(err, ShouldBeNil)

				srcA, isMap := src["a"].(map[string]interface{})
				So(isMap, ShouldBeTrue)

				srcB, isMap := srcA["b"].(map[string]interface{})
				So(isMap, ShouldBeTrue)

				srcD, isMap := srcB["d"].(map[string]interface{})
				So(isMap, ShouldBeTrue)

				copy := deepCopyMap(src)

				copyA, isMap := copy["a"].(map[string]interface{})
				So(isMap, ShouldBeTrue)

				copyB, isMap := copyA["b"].(map[string]interface{})
				So(isMap, ShouldBeTrue)

				copyD, isMap := copyB["d"].(map[string]interface{})
				So(isMap, ShouldBeTrue)

				//copy and src should be equal
				So(srcB["c"], ShouldEqual, 100)
				So(copyB["c"], ShouldEqual, 100)

				So(srcD["e"], ShouldEqual, "200")
				So(copyD["e"], ShouldEqual, "200")

				//once we modify the copy...
				newD := map[string]interface{}{}
				newD["e"] = 500
				copyB["d"] = newD

				//source shoud stay the same, copy should be changed
				srcD, _ = srcB["d"].(map[string]interface{})
				copyD, _ = copyB["d"].(map[string]interface{})
				So(srcD["e"], ShouldEqual, "200")
				So(copyD["e"], ShouldEqual, 500)
			})
		})

		Convey("findOverrideValue function", func() {

			Convey("Should find non-empty value in a map", func() {
				flatmap := map[string]string{}
				flatmap["a.b.c.d"] = "testval"

				oMap := UnflattenToMap(flatmap)

				val, exists := FindOverrideStringValue(oMap, "a.b.c.d")
				So(exists, ShouldBeTrue)
				So(val, ShouldEqual, "testval")
			})

			Convey("Should find empty string in a map", func() {
				flatmap := map[string]string{}
				flatmap["a.b.c.d"] = ""

				oMap := UnflattenToMap(flatmap)

				val, exists := FindOverrideStringValue(oMap, "a.b.c.d")
				So(exists, ShouldBeTrue)
				So(val, ShouldBeBlank)
			})

			Convey("Should not find override value in a map when it's not a final entry", func() {
				flatmap := map[string]string{}
				flatmap["a.b.c.d"] = "testval"

				oMap := UnflattenToMap(flatmap)

				_, exists := FindOverrideStringValue(oMap, "a.b.c")
				So(exists, ShouldBeFalse)
			})

			Convey("Should not find override value in a map when it does not exist", func() {
				flatmap := map[string]string{}
				flatmap["a.b.c.d"] = "testval"

				oMap := UnflattenToMap(flatmap)

				_, exists := FindOverrideStringValue(oMap, "a.b.f")
				So(exists, ShouldBeFalse)
			})
		})

		Convey("UnflattenToMap function", func() {

			Convey("Should unflatten the map", func() {
				flatmap := map[string]string{}
				flatmap["a.b.c"] = "testval"

				oMap := UnflattenToMap(flatmap)

				val, exists := FindOverrideStringValue(oMap, "a.b.c")
				So(exists, ShouldBeTrue)
				So(val, ShouldEqual, "testval")
			})

			Convey("Should convert any strings that contain booleans to booleans", func() {
				flatmap := map[string]string{}
				flatmap["a.b.c"] = "true"
				flatmap["a.b.d"] = "false"

				oMap := UnflattenToMap(flatmap)

				val := FindOverrideValue(oMap, "a.b.c")
				val2 := FindOverrideValue(oMap, "a.b.d")

				So(val, ShouldBeTrue)
				So(val2, ShouldBeFalse)
			})
		})
	})
}

//nlnl == [n]o [l]eading [n]ew [l]ine
func nlnl(s string) string {
	return strings.TrimLeft(s, "\n")
}
