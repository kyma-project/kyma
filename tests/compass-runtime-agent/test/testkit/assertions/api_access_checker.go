package assertions

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

// TODO: we should consider enhancing test with sending events (also use Mock Service)

type APIAccessChecker struct {
}

func NewAPIAccessChecker() *APIAccessChecker {
	return &APIAccessChecker{}
}

func (c *APIAccessChecker) AssertAPIAccess(t *testing.T, apis []*graphql.APIDefinition) {
	// TODO - call access service on correct path based on credentials? or only status/ok and do different tests proxy specific?
}
