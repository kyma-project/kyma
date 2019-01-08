package test

import (
	"github.com/kyma-project/kyma/tools/failery/failery/fixtures/http"
)

type HasConflictingNestedImports interface {
	RequesterNS
	Z() http.MyStruct
}
