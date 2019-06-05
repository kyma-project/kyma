package handlers

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func Test_filterCEHeaders(t *testing.T) {
	headers := make(map[string][]string)
	request := http.Request{
		Header: headers,
	}
	result := filterCEHeaders(&request)
	assert.Len(t, result, 0)

	headers["ce-test-header"] = []string{"test-value"}
	result = filterCEHeaders(&request)
	assert.Len(t, result, 1)
	assert.Equal(t, headers["ce-test-header"][0], "test-value")

	headers["NO-ce-test-header"] = []string{"NO-test-value"}
	result = filterCEHeaders(&request)
	assert.Len(t, result, 1)
	assert.Equal(t, headers["ce-test-header"][0], "test-value")
}
