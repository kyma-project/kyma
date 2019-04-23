// +build acceptance

package k8s

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/dex"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/auth"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type selfSubjectRulesQueryResponse struct {
	SelfSubjectRules selfSubjectRules `json:"selfSubjectRules"`
}

type selfSubjectRules struct {
	ResourceRules []*resourceRule `json:"resourceRules"`
}

type resourceRule struct {
	Verbs     []string `json:"verbs"`
	APIGroups []string `json:"apiGroups"`
	Resources []string `json:"resources"`
}

func TestSelfSubjectRules(t *testing.T) {
	dex.SkipTestIfSCIEnabled(t)

	c, err := graphql.New()
	require.NoError(t, err)

	t.Log("Querying for SelfSubjctRules...")

	var selfSubjectRulesRes selfSubjectRulesQueryResponse

	err = c.Do(fixSelfSubjectRulesQuery(), &selfSubjectRulesRes)
	require.NoError(t, err)
	assert.True(t, len(selfSubjectRulesRes.SelfSubjectRules.ResourceRules) > 0)

	err = c.Do(fixNamespacedSelfSubjectRulesQuery("foo"), &selfSubjectRulesRes)
	require.NoError(t, err)
	assert.True(t, len(selfSubjectRulesRes.SelfSubjectRules.ResourceRules) > 0)

	t.Log("Checking authorization directives...")
	ops := &auth.OperationsInput{
		auth.CreateSelfSubjectRulesReview: {fixSelfSubjectRulesQuery()},
	}
	AuthSuite.Run(t, ops)
}

func fixSelfSubjectRulesQuery() *graphql.Request {
	query := fmt.Sprintf(
		`query {
		selfSubjectRules {
			resourceRules{
				verbs
				resources
				apiGroups
			}
		}
	}`)
	return graphql.NewRequest(query)
}

func fixNamespacedSelfSubjectRulesQuery(namespace string) *graphql.Request {
	query := fmt.Sprintf(
		`query ($namespace: String){
		selfSubjectRules (namespace: $namespace){
			resourceRules{
				verbs
				resources
				apiGroups
			}
		}
	}`)
	req := graphql.NewRequest(query)
	req.SetVar("namespace", namespace)
	return req
}
