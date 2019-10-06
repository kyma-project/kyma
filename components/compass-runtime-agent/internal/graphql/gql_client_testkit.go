package graphql

import (
	"errors"
	"testing"

	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
)

type QueryAssertClient struct {
	t                  *testing.T
	expectedRequest    *graphql.Request
	shouldFail         bool
	modifyResponseFunc ModifyResponseFunc
}

type ModifyResponseFunc func(t *testing.T, r interface{})

func (c *QueryAssertClient) Do(req *graphql.Request, res interface{}) error {
	assert.Equal(c.t, c.expectedRequest, req)

	if !c.shouldFail {
		c.modifyResponseFunc(c.t, res)

		return nil
	}

	return errors.New("error")
}

func NewQueryAssertClient(t *testing.T, expectedReq *graphql.Request, shouldFail bool, modifyResponseFunc func(t *testing.T, r interface{})) Client {
	return &QueryAssertClient{
		t:                  t,
		expectedRequest:    expectedReq,
		shouldFail:         shouldFail,
		modifyResponseFunc: modifyResponseFunc,
	}
}
