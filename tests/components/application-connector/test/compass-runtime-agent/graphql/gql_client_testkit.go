package graphql

import (
	"errors"
	"testing"

	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/third_party/machinebox/graphql"
	"github.com/stretchr/testify/assert"
)

type QueryAssertClient struct {
	t                  *testing.T
	expectedRequests   []*graphql.Request
	err                error
	modifyResponseFunc ModifyResponseFunc
}

type ModifyResponseFunc []func(t *testing.T, r interface{})

func (c *QueryAssertClient) Do(req *graphql.Request, res interface{}) error {
	if len(c.expectedRequests) == 0 {
		return errors.New("no more requests were expected")
	}

	assert.Equal(c.t, c.expectedRequests[0], req)
	if len(c.expectedRequests) > 1 {
		c.expectedRequests = c.expectedRequests[1:]
	}

	if len(c.modifyResponseFunc) > 0 {
		c.modifyResponseFunc[0](c.t, res)
		if len(c.modifyResponseFunc) > 1 {
			c.modifyResponseFunc = c.modifyResponseFunc[1:]
		}
	}

	return c.err
}

func NewQueryAssertClient(t *testing.T, err error, expectedReq []*graphql.Request, modifyResponseFunc ...func(t *testing.T, r interface{})) Client {
	return &QueryAssertClient{
		t:                  t,
		expectedRequests:   expectedReq,
		err:                err,
		modifyResponseFunc: modifyResponseFunc,
	}
}
