// Package ptr provides conversion rules to pointers for DTO construction.
// Rules taken from AWS go SDK package on Apache License, Version 2.0 licence.
package ptr

import "time"

// String returns a pointer to the string value passed in.
func String(v string) *string {
	return &v
}

// Bool returns a pointer to the bool value passed in.
func Bool(v bool) *bool {
	return &v
}

// Int returns a pointer to the int value passed in.
func Int(v int) *int {
	return &v
}

// Int64 returns a pointer to the int64 value passed in.
func Int64(val int64) *int64 {
	return &val
}

// Float32 returns a pointer to the float32 value passed in.
func Float32(v float32) *float32 {
	return &v
}

// Float64 returns a pointer to the float64 value passed in.
func Float64(v float64) *float64 {
	return &v
}

// Time returns a pointer to the time.Time value passed in.
func Time(v time.Time) *time.Time {
	return &v
}
