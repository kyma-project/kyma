package testing

import (
	"fmt"
	"time"

	"github.com/kyma-project/kyma/components/helm-broker/internal/storage"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage/driver/etcd"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage/driver/memory"
)

func mustInstanceOperationWithClock(u storage.InstanceOperation, nowProvider func() time.Time) storage.InstanceOperation {
	switch uCst := u.(type) {
	case *memory.InstanceOperation:
		return uCst.WithTimeProvider(nowProvider)
	case *etcd.InstanceOperation:
		return uCst.WithTimeProvider(nowProvider)
	default:
	}

	panic(fmt.Sprintf("unsupported InstanceOperation storage type: %T", u))
}
