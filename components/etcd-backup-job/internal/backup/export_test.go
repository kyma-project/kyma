package backup

import "github.com/kyma-project/kyma/components/etcd-backup-job/internal/platform/idprovider"

func (e *Executor) WithIdProvider(idProvider idprovider.Fn) *Executor{
	e.idProvider = idProvider
	return e
}
