package cms

import (
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
)

type notifier interface {
	AddListener(observer resource.Listener)
	DeleteListener(observer resource.Listener)
}
