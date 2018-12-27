package test

import (
	"net/http"

	my_http "github.com/kyma-project/kyma/tools/failery/failery/fixtures/http"
)

// Example is an example
type Example interface {
	A() http.Flusher
	B(fixtureshttp string) my_http.MyStruct
}
