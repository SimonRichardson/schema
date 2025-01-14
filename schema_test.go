// Copyright 2011, 2012, 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package schema_test

import (
	"fmt"
	"math"
	"net/url"
	"time"

	gc "gopkg.in/check.v1"

	"github.com/juju/schema"
)

var aPath = []string{"<pa", "th>"}

type Dummy struct{}

func (d *Dummy) Coerce(value interface{}, path []string) (coerced interface{}, err error) {
	return "i-am-dummy", nil
}

type S struct{}

var _ = gc.Suite(&S{})

func (s *S) TestConst(c *gc.C) {
	sch := schema.Const("foo")

	out, err := sch.Coerce("foo", aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, "foo")

	out, err = sch.Coerce(42, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `<path>: expected "foo", got int\(42\)`)

	out, err = sch.Coerce(nil, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `<path>: expected "foo", got nothing`)
}

func (s *S) TestNilSuccess(c *gc.C) {
	assertSuccess := func(sch schema.Checker) {
		out, err := sch.Coerce(nil, aPath)
		c.Assert(err, gc.IsNil)
		c.Assert(out, gc.IsNil)
	}

	sch := schema.Nil("")
	assertSuccess(sch)

	sch = schema.Nil("any label")
	assertSuccess(sch)
}

var nonNilValues = []interface{}{42, "", "foo", false, 3.14, 0}

func (s *S) TestNilFailuresWithEmptyLabel(c *gc.C) {
	sch := schema.Nil("")
	testCheckerFailsForEachBadValueWithErrorPrefix(sch, c, nonNilValues, `<path>: expected empty value`)
}

func (s *S) TestNilFailuresWithNonEmptyLabel(c *gc.C) {
	sch := schema.Nil("wallet")
	testCheckerFailsForEachBadValueWithErrorPrefix(sch, c, nonNilValues, `<path>: expected empty wallet`)
}

func (s *S) TestNonEmptyStringSuccess(c *gc.C) {
	assertSuccess := func(sch schema.Checker) {
		out, err := sch.Coerce("non-empty value is ok", aPath)
		c.Assert(err, gc.IsNil)
		c.Assert(out, gc.Equals, "non-empty value is ok")
	}

	sch := schema.NonEmptyString("")
	assertSuccess(sch)

	sch = schema.NonEmptyString("any label")
	assertSuccess(sch)
}

var badNonEmptyStringValues = []interface{}{42, "", []string{"x"}, false, 3.14, 0}

func (s *S) TestNonEmptyStringWithEmptyLabel(c *gc.C) {
	sch := schema.NonEmptyString("")
	testCheckerFailsForEachBadValueWithErrorPrefix(sch, c, badNonEmptyStringValues, `<path>: expected non-empty string`)
}

func (s *S) TestNonEmptyStringWithNonEmptyLabel(c *gc.C) {
	sch := schema.NonEmptyString("free beer")
	testCheckerFailsForEachBadValueWithErrorPrefix(sch, c, badNonEmptyStringValues, `<path>: expected non-empty free beer`)
}

func testCheckerFailsForEachBadValueWithErrorPrefix(sch schema.Checker, c *gc.C, badValues []interface{}, errorPrefix string) {
	for _, badValue := range badValues {
		out, err := sch.Coerce(badValue, aPath)
		c.Check(out, gc.IsNil)
		errorSuffix := fmt.Sprintf(", got %T(%#v)", badValue, badValue)
		c.Check(err, gc.NotNil)
		c.Check(err.Error(), gc.Equals, errorPrefix+errorSuffix)
	}
}

func (s *S) TestAny(c *gc.C) {
	sch := schema.Any()

	out, err := sch.Coerce("foo", aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, "foo")

	out, err = sch.Coerce(nil, aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, nil)
}

func (s *S) TestOneOf(c *gc.C) {
	sch := schema.OneOf(schema.Const("foo"), schema.Const(42))

	out, err := sch.Coerce("foo", aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, "foo")

	out, err = sch.Coerce(42, aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, 42)

	out, err = sch.Coerce("bar", aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `<path>: unexpected value "bar"`)
}

func (s *S) TestBool(c *gc.C) {
	sch := schema.Bool()

	for _, trueValue := range []interface{}{true, "1", "true", "True", "TRUE"} {
		out, err := sch.Coerce(trueValue, aPath)
		c.Assert(err, gc.IsNil)
		c.Assert(out, gc.Equals, true)
	}

	for _, falseValue := range []interface{}{false, "0", "false", "False", "FALSE"} {
		out, err := sch.Coerce(falseValue, aPath)
		c.Assert(err, gc.IsNil)
		c.Assert(out, gc.Equals, false)
	}

	out, err := sch.Coerce(42, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `<path>: expected bool, got int\(42\)`)

	out, err = sch.Coerce(nil, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, "<path>: expected bool, got nothing")
}

func (s *S) TestInt(c *gc.C) {
	sch := schema.Int()

	out, err := sch.Coerce(42, aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, int64(42))

	out, err = sch.Coerce(int8(42), aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, int64(42))

	out, err = sch.Coerce("42", aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, int64(42))

	out, err = sch.Coerce(true, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `<path>: expected int, got bool\(true\)`)

	out, err = sch.Coerce(nil, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, "<path>: expected int, got nothing")
}

func (s *S) TestUint(c *gc.C) {
	sch := schema.Uint()

	out, err := sch.Coerce(42, aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, uint64(42))

	out, err = sch.Coerce(int8(42), aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, uint64(42))

	out, err = sch.Coerce(uint8(42), aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, uint64(42))

	out, err = sch.Coerce("42", aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, uint64(42))

	out, err = sch.Coerce("-42", aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err.Error(), gc.Equals, `<path>: expected uint, got string("-42")`)

	out, err = sch.Coerce(-42, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err.Error(), gc.Equals, "<path>: expected uint, got int(-42)")

	out, err = sch.Coerce(true, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err.Error(), gc.Equals, "<path>: expected uint, got bool(true)")

	out, err = sch.Coerce(nil, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err.Error(), gc.Equals, "<path>: expected uint, got nothing")
}

func (s *S) TestForceInt(c *gc.C) {
	sch := schema.ForceInt()

	out, err := sch.Coerce(42, aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, int(42))

	out, err = sch.Coerce("42", aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, int(42))

	out, err = sch.Coerce("42.66", aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, int(42))

	out, err = sch.Coerce(int8(42), aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, int(42))

	out, err = sch.Coerce(float32(42), aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, int(42))

	out, err = sch.Coerce(float64(42), aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, int(42))

	out, err = sch.Coerce(42.66, aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, int(42))

	// If an out of range value is provided, that value is truncated,
	// generating unexpected results, but no error is raised.
	out, err = sch.Coerce(float64(math.MaxInt64+1), aPath)
	c.Assert(err, gc.IsNil)

	out, err = sch.Coerce(true, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `<path>: expected number, got bool\(true\)`)

	out, err = sch.Coerce(nil, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, "<path>: expected number, got nothing")
}

func (s *S) TestForceUint(c *gc.C) {
	sch := schema.ForceUint()

	out, err := sch.Coerce(42, aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, uint64(42))

	out, err = sch.Coerce(-42, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err.Error(), gc.Equals, `<path>: expected uint, got int(-42)`)

	out, err = sch.Coerce("-42", aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err.Error(), gc.Equals, `<path>: expected uint, got string("-42")`)

	out, err = sch.Coerce("42", aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, uint64(42))

	out, err = sch.Coerce("42.66", aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, uint64(42))

	out, err = sch.Coerce(int8(42), aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, uint64(42))

	out, err = sch.Coerce(uint8(42), aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, uint64(42))

	out, err = sch.Coerce(float32(42), aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, uint64(42))

	out, err = sch.Coerce(float64(42), aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, uint64(42))

	out, err = sch.Coerce(42.66, aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, uint64(42))

	out, err = sch.Coerce(float64(-42), aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err.Error(), gc.Equals, `<path>: expected uint, got float64(-42)`)

	// If an out of range value is provided, that value is truncated,
	// generating unexpected results, but no error is raised.
	out, err = sch.Coerce(float64(math.MaxInt64+1), aPath)
	c.Assert(err, gc.IsNil)

	out, err = sch.Coerce(true, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `<path>: expected uint, got bool\(true\)`)

	out, err = sch.Coerce(nil, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, "<path>: expected uint, got nothing")
}

func (s *S) TestFloat(c *gc.C) {
	sch := schema.Float()

	out, err := sch.Coerce(float32(1.0), aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, float64(1.0))

	out, err = sch.Coerce(float64(1.0), aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, float64(1.0))

	out, err = sch.Coerce(int8(1), aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, float64(1))

	out, err = sch.Coerce(int64(math.MaxInt64), aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, float64(math.MaxInt64))

	out, err = sch.Coerce(int64(math.MinInt64), aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, float64(math.MinInt64))

	out, err = sch.Coerce(uint64(0), aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, float64(0))

	out, err = sch.Coerce(uint64(math.MaxUint64), aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, float64(math.MaxUint64))

	out, err = sch.Coerce(true, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `<path>: expected float, got bool\(true\)`)

	out, err = sch.Coerce(nil, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, "<path>: expected float, got nothing")
}

func (s *S) TestString(c *gc.C) {
	sch := schema.String()

	out, err := sch.Coerce("foo", aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, "foo")

	out, err = sch.Coerce(true, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `<path>: expected string, got bool\(true\)`)

	out, err = sch.Coerce(nil, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, "<path>: expected string, got nothing")
}

func (s *S) TestURL(c *gc.C) {
	mustParse := func(s string) *url.URL {
		u, err := url.Parse(s)
		c.Assert(err, gc.IsNil)
		return u
	}
	sch := schema.URL()

	out, err := sch.Coerce("foo", aPath)
	c.Assert(err, gc.IsNil)
	expected := mustParse("foo")
	c.Assert(*(out.(*url.URL)), gc.Equals, *expected)

	out, err = sch.Coerce(true, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `<path>: expected url string, got bool\(true\)`)

	out, err = sch.Coerce(nil, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, "<path>: expected url string, got nothing")

	out, err = sch.Coerce(":::", aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `<path>: expected valid url, got string\(":::"\)`)
}

func (s *S) TestSimpleRegexp(c *gc.C) {
	sch := schema.SimpleRegexp()
	out, err := sch.Coerce("[0-9]+", aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, "[0-9]+")

	out, err = sch.Coerce(1, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `<path>: expected regexp string, got int\(1\)`)

	out, err = sch.Coerce("[", aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `<path>: expected valid regexp, got string\("\["\)`)

	out, err = sch.Coerce(nil, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `<path>: expected regexp string, got nothing`)
}

func (s *S) TestList(c *gc.C) {
	sch := schema.List(schema.Int())
	out, err := sch.Coerce([]int8{1, 2}, aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.DeepEquals, []interface{}{int64(1), int64(2)})

	out, err = sch.Coerce(42, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, "<path>: expected list, got int\\(42\\)")

	out, err = sch.Coerce(nil, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, "<path>: expected list, got nothing")

	out, err = sch.Coerce([]interface{}{1, true}, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `<path>\[1\]: expected int, got bool\(true\)`)
}

func (s *S) TestMap(c *gc.C) {
	sch := schema.Map(schema.String(), schema.Int())
	out, err := sch.Coerce(map[string]interface{}{"a": 1, "b": int8(2)}, aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.DeepEquals, map[interface{}]interface{}{"a": int64(1), "b": int64(2)})

	out, err = sch.Coerce(42, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, "<path>: expected map, got int\\(42\\)")

	out, err = sch.Coerce(nil, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, "<path>: expected map, got nothing")

	out, err = sch.Coerce(map[int]int{1: 1}, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, "<path>: expected string, got int\\(1\\)")

	out, err = sch.Coerce(map[string]bool{"a": true}, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `<path>\.a: expected int, got bool\(true\)`)

	// First path entry shouldn't have dots in an error message.
	out, err = sch.Coerce(map[string]bool{"a": true}, nil)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `a: expected int, got bool\(true\)`)

	// Error should work even when path is nil.
	out, err = sch.Coerce(nil, nil)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `expected map, got nothing`)
}

func (s *S) TestStringMap(c *gc.C) {
	sch := schema.StringMap(schema.Int())
	out, err := sch.Coerce(map[string]interface{}{"a": 1, "b": int8(2)}, aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.DeepEquals, map[string]interface{}{"a": int64(1), "b": int64(2)})

	out, err = sch.Coerce(42, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, "<path>: expected map, got int\\(42\\)")

	out, err = sch.Coerce(nil, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, "<path>: expected map, got nothing")

	out, err = sch.Coerce(map[int]int{1: 1}, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, "<path>: expected string, got int\\(1\\)")

	out, err = sch.Coerce(map[string]bool{"a": true}, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `<path>\.a: expected int, got bool\(true\)`)

	// First path entry shouldn't have dots in an error message.
	out, err = sch.Coerce(map[string]bool{"a": true}, nil)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `a: expected int, got bool\(true\)`)
}

func assertFieldMap(c *gc.C, sch schema.Checker) {
	out, err := sch.Coerce(map[string]interface{}{"a": "A", "b": "B"}, aPath)

	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.DeepEquals, map[string]interface{}{"a": "A", "b": "B", "c": "C"})

	out, err = sch.Coerce(42, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, "<path>: expected map, got int\\(42\\)")

	out, err = sch.Coerce(nil, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, "<path>: expected map, got nothing")

	out, err = sch.Coerce(map[string]interface{}{"a": "A", "b": "C"}, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `<path>\.b: expected "B", got string\("C"\)`)

	out, err = sch.Coerce(map[string]interface{}{"b": "B"}, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `<path>\.a: expected "A", got nothing`)

	// b is optional
	out, err = sch.Coerce(map[string]interface{}{"a": "A"}, aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.DeepEquals, map[string]interface{}{"a": "A", "c": "C"})

	// First path entry shouldn't have dots in an error message.
	out, err = sch.Coerce(map[string]bool{"a": true}, nil)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `a: expected "A", got bool\(true\)`)
}

func (s *S) TestFieldMapBasic(c *gc.C) {
	fields := schema.Fields{
		"a": schema.Const("A"),
		"b": schema.Const("B"),
		"c": schema.Const("C"),
	}
	defaults := schema.Defaults{
		"b": schema.Omit,
		"c": "C",
	}
	sch := schema.FieldMap(fields, defaults)
	assertFieldMap(c, sch)

	out, err := sch.Coerce(map[string]interface{}{"a": "A", "b": "B", "d": "D"}, aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.DeepEquals, map[string]interface{}{"a": "A", "b": "B", "c": "C"})

	out, err = sch.Coerce(map[string]interface{}{"a": "A", "d": "D"}, aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.DeepEquals, map[string]interface{}{"a": "A", "c": "C"})

	out, err = sch.Coerce(123, aPath)
	c.Assert(err, gc.ErrorMatches, `<path>: expected map, got int\(123\)`)
	c.Assert(out, gc.Equals, nil)

	out, err = sch.Coerce(map[int]string{}, aPath)
	c.Assert(err, gc.ErrorMatches, `<path>: expected map\[string], got map\[int]string\(map\[int]string{}\)`)
	c.Assert(out, gc.Equals, nil)

	type strKey string
	out, err = sch.Coerce(map[strKey]string{"a": "A"}, aPath)
	c.Assert(err, gc.ErrorMatches, `<path>: expected map\[string], got map\[schema_test\.strKey]string\(map\[schema_test.strKey]string{"a":"A"}\)`)
	c.Assert(out, gc.Equals, nil)
}

func (s *S) TestFieldMapInterfaceKey(c *gc.C) {
	fields := schema.Fields{
		"a": schema.Const("A"),
	}
	sch := schema.FieldMap(fields, nil)

	out, err := sch.Coerce(map[interface{}]interface{}{"a": "A"}, aPath)
	c.Assert(err, gc.IsNil)
	c.Check(out, gc.DeepEquals, map[string]interface{}{"a": "A"})

	_, err = sch.Coerce(map[interface{}]interface{}{1: "A"}, aPath)
	c.Check(err, gc.ErrorMatches, `<path>: expected map\[string], got map\[interface {}]interface {}\(map\[interface {}]interface {}{1:"A"}\)`)
}

func (s *S) TestFieldMapDefaultInvalid(c *gc.C) {
	fields := schema.Fields{
		"a": schema.Const("A"),
	}
	defaults := schema.Defaults{
		"a": "B",
	}
	sch := schema.FieldMap(fields, defaults)
	_, err := sch.Coerce(map[string]interface{}{}, aPath)
	c.Assert(err, gc.ErrorMatches, `<path>.a: expected "A", got string\("B"\)`)
}

func (s *S) TestStrictFieldMap(c *gc.C) {
	fields := schema.Fields{
		"a": schema.Const("A"),
		"b": schema.Const("B"),
		"c": schema.Const("C"),
	}
	defaults := schema.Defaults{
		"b": schema.Omit,
		"c": "C",
	}
	sch := schema.StrictFieldMap(fields, defaults)
	assertFieldMap(c, sch)

	out, err := sch.Coerce(map[string]interface{}{"a": "A", "b": "B", "d": "D"}, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `<path>: unknown key "d" \(value "D"\)`)

	out, err = sch.Coerce(map[string]interface{}{"a": "A", "b": "B", "d": "D"}, nil)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `unknown key "d" \(value "D"\)`)
}

func (s *S) TestSchemaMap(c *gc.C) {
	fields1 := schema.FieldMap(schema.Fields{
		"type": schema.Const(1),
		"a":    schema.Const(2),
	}, nil)
	fields2 := schema.FieldMap(schema.Fields{
		"type": schema.Const(3),
		"b":    schema.Const(4),
	}, nil)
	sch := schema.FieldMapSet("type", []schema.Checker{fields1, fields2})

	out, err := sch.Coerce(map[string]int{"type": 1, "a": 2}, aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.DeepEquals, map[string]interface{}{"type": 1, "a": 2})

	out, err = sch.Coerce(map[string]int{"type": 3, "b": 4}, aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.DeepEquals, map[string]interface{}{"type": 3, "b": 4})

	out, err = sch.Coerce(map[string]int{}, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `<path>\.type: expected supported selector, got nothing`)

	out, err = sch.Coerce(map[string]int{"type": 2}, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `<path>\.type: expected supported selector, got int\(2\)`)

	out, err = sch.Coerce(map[string]int{"type": 3, "b": 5}, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `<path>\.b: expected 4, got int\(5\)`)

	out, err = sch.Coerce(42, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `<path>: expected map, got int\(42\)`)

	out, err = sch.Coerce(nil, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `<path>: expected map, got nothing`)

	// First path entry shouldn't have dots in an error message.
	out, err = sch.Coerce(map[string]int{"a": 1}, nil)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `type: expected supported selector, got nothing`)
}

func (s *S) TestUUID(c *gc.C) {
	sch := schema.UUID()

	out, err := sch.Coerce("6216dfc3-6e82-408f-9f74-8565e63e6158", aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, "6216dfc3-6e82-408f-9f74-8565e63e6158")

	out, err = sch.Coerce("uuid", aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, `<path>: expected uuid, got string\(\"uuid\"\)`)

	out, err = sch.Coerce(nil, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, "<path>: expected uuid, got nothing")
}

func (s *S) TestTime(c *gc.C) {
	sch := schema.Time()

	var empty time.Time
	value := time.Date(2016, 10, 9, 12, 34, 56, 0, time.UTC)

	out, err := sch.Coerce("", aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, empty)

	out, err = sch.Coerce(value.Format(time.RFC3339Nano), aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, value)

	out, err = sch.Coerce("invalid", aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err.Error(), gc.Equals, `<path>: conversion to time: parsing time "invalid" as "2006-01-02T15:04:05.999999999Z07:00": cannot parse "invalid" as "2006"`)

	out, err = sch.Coerce(42, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err.Error(), gc.Equals, "<path>: expected string or time.Time, got int(42)")

	out, err = sch.Coerce(nil, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err.Error(), gc.Equals, "<path>: expected string or time.Time, got nothing")
}

func (s *S) TestStringified(c *gc.C) {
	sch := schema.Stringified()

	out, err := sch.Coerce(true, aPath)
	c.Assert(err, gc.IsNil)
	c.Check(out, gc.Equals, "true")

	out, err = sch.Coerce(10, aPath)
	c.Assert(err, gc.IsNil)
	c.Check(out, gc.Equals, "10")

	out, err = sch.Coerce(1.1, aPath)
	c.Assert(err, gc.IsNil)
	c.Check(out, gc.Equals, "1.1")

	out, err = sch.Coerce("spam", aPath)
	c.Assert(err, gc.IsNil)
	c.Check(out, gc.Equals, "spam")

	_, err = sch.Coerce(map[string]string{}, aPath)
	c.Check(err, gc.ErrorMatches, ".* unexpected value .*")

	_, err = sch.Coerce([]string{}, aPath)
	c.Check(err, gc.ErrorMatches, ".* unexpected value .*")

	sch = schema.Stringified(schema.StringMap(schema.String()))

	out, err = sch.Coerce(map[string]string{"a": "b"}, aPath)
	c.Assert(err, gc.IsNil)
	c.Check(out, gc.Equals, `map[string]string{"a":"b"}`)
}

func (s *S) TestTimeDuration(c *gc.C) {
	sch := schema.TimeDuration()

	var empty time.Duration

	out, err := sch.Coerce("", aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, empty)

	value, _ := time.ParseDuration("18h")

	out, err = sch.Coerce("18h", aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, value)

	out, err = sch.Coerce("failure", aPath)
	c.Assert(err.Error(), gc.Equals, "<path>: conversion to duration: time: invalid duration \"failure\"")

	out, err = sch.Coerce(42, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err.Error(), gc.Equals, "<path>: expected string or time.Duration, got int(42)")

	out, err = sch.Coerce(nil, aPath)
	c.Assert(out, gc.IsNil)
	c.Assert(err.Error(), gc.Equals, "<path>: expected string or time.Duration, got nothing")
}

func (s *S) TestSize(c *gc.C) {
	sch := schema.Size()
	//Invalid size
	out, err := sch.Coerce("18X", aPath)
	c.Assert(err.Error(), gc.Equals, "invalid multiplier suffix \"X\", expected one of MGTPEZY")
	c.Assert(out, gc.IsNil)

	//Valid Size
	out, err = sch.Coerce("18G", aPath)
	c.Assert(err, gc.IsNil)
	c.Assert(out, gc.Equals, uint64(18432))

	//Empty string
	out, err = sch.Coerce("", aPath)
	c.Assert(err.Error(), gc.Equals, "<path>: expected empty string, got string(\"\")")
	c.Assert(out, gc.IsNil)

	//Nil
	out, err = sch.Coerce(nil, aPath)
	c.Assert(err.Error(), gc.Equals, "<path>: expected string, got nothing")
	c.Assert(out, gc.IsNil)

	//Invalid type
	var foo int
	out, err = sch.Coerce(foo, aPath)
	c.Assert(err.Error(), gc.Equals, "<path>: expected string, got int(0)")
	c.Assert(out, gc.IsNil)
}
