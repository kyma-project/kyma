package graphql

import (
	"errors"
	"testing"

	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
)

type QueryAssertClient struct {
	t             *testing.T
	responseMocks []ResponseMock
	shouldFail    bool
}

type ModifyResponseFunc func(t *testing.T, r interface{})

type ResponseMock struct {
	ModifyResponseFunc func(t *testing.T, r interface{})
	ExpectedReq        *graphql.Request
}

func (c *QueryAssertClient) Do(req *graphql.Request, res interface{}) error {
	if len(c.responseMocks) == 0 {
		return errors.New("no more requests were expected")
	}

	currentResponseMock := c.responseMocks[0]

	assert.Equal(c.t, currentResponseMock.ExpectedReq, req)
	if len(c.responseMocks) > 1 {
		c.responseMocks = c.responseMocks[1:]
	}

	if !c.shouldFail {
		currentResponseMock.ModifyResponseFunc(c.t, res)

		return nil
	}

	return errors.New("error")
}

func NewQueryAssertClient(t *testing.T, shouldFail bool, responseMocks ...ResponseMock) Client {
	return &QueryAssertClient{
		t:             t,
		responseMocks: responseMocks,
		shouldFail:    shouldFail,
	}
}
