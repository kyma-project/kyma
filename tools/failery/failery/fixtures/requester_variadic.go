package test

import "io"

type RequesterVariadic interface {
	// cases: only variadic argument, w/ and w/out interface type
	Get(values ...string) (bool, *error)
	OneInterface(a ...interface{}) (bool, error)

	// cases: normal argument + variadic argument, w/ and w/o interface type
	Sprintf(format string, a ...interface{}) (string, error)
	MultiWriteToFile(filename string, w ...io.Writer) (string, error)
}
