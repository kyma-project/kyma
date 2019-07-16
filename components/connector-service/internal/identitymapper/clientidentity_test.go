package identitymapper

import (
	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	runtimeID   = "runtimeID"
	application = "application"
	group       = "group"
	tenant      = "tenant"
)

func TestToApplicationIdentity(t *testing.T) {
	//given
	clientCxt := clientcontext.ClientContext{
		ClusterContext: clientcontext.ClusterContext{
			Tenant: tenant,
			Group:  group,
		},
		ID: application,
	}

	//when
	identity := ToApplicationIdentity(clientCxt)

	//then
	applicationIdentity := identity.(ApplicationIdentity)
	assert.Equal(t, application, applicationIdentity.Application)
	assert.Equal(t, tenant, applicationIdentity.Tenant)
	assert.Equal(t, group, applicationIdentity.Group)
}

func TestToRuntimeIdentity(t *testing.T) {
	//given
	clientCxt := clientcontext.ClientContext{
		ClusterContext: clientcontext.ClusterContext{
			Tenant: tenant,
			Group:  group,
		},
		ID: runtimeID,
	}

	//when
	identity := ToRuntimeIdentity(clientCxt)

	//then
	runtimeIdentity := identity.(RuntimeIdentity)
	assert.Equal(t, runtimeID, runtimeIdentity.RuntimeID)
	assert.Equal(t, tenant, runtimeIdentity.Tenant)
	assert.Equal(t, group, runtimeIdentity.Group)
}
