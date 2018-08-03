package backup

import "github.com/kyma-project/kyma/tools/etcd-backup/internal/platform/idprovider"

func (e *Executor) WithIdProvider(idProvider idprovider.Fn) *Executor{
	e.idProvider = idProvider
	return e
}
