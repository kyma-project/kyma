package graphql

import (
	"errors"
	"testing"

	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
)

type QueryAssertClient struct {
	t                  *testing.T
	expectedRequests   []*graphql.Request
	shouldFail         bool
	modifyResponseFunc ModifyResponseFunc
}

type ModifyResponseFunc func(t *testing.T, r interface{})

func (c *QueryAssertClient) Do(req *graphql.Request, res interface{}) error {
	if len(c.expectedRequests) == 0 {
		return errors.New("no more requests were expected")
	}

	assert.Equal(c.t, c.expectedRequests[0], req)
	if len(c.expectedRequests) > 1 {
		c.expectedRequests = c.expectedRequests[1:]
	}

	if !c.shouldFail {
		c.modifyResponseFunc(c.t, res)

		return nil
	}

	return errors.New("error")
}

func NewQueryAssertClient(t *testing.T, shouldFail bool, modifyResponseFunc func(t *testing.T, r interface{}), expectedReq ...*graphql.Request) Client {
	return &QueryAssertClient{
		t:                  t,
		expectedRequests:   expectedReq,
		shouldFail:         shouldFail,
		modifyResponseFunc: modifyResponseFunc,
	}
}
