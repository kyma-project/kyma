package cloudevent_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/cloudevent"
	"github.com/stretchr/testify/assert"
)

func TestNewHTTP(t *testing.T) {
	// SuT: ClientFactory
	// UoW: NewHTTP()
	// test kind: value

	cf := cloudevent.ClientFactory{}
	client, err := cf.NewHTTP()
	assert.NoError(t, err)
	assert.NotNil(t, client)
}
