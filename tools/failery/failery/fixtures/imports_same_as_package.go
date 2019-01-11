package test

import (
	"github.com/kyma-project/kyma/tools/failery/failery/fixtures/test"
)

type C int

type ImportsSameAsPackage interface {
	A() test.B
	B() KeyManager
	C(C)
}
