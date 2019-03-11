package content

import (
	"fmt"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/storage"
)

type odataSpecService struct {
	storage minioODataSpecGetter
}

func newODataSpecService(storage minioODataSpecGetter) *odataSpecService {
	return &odataSpecService{
		storage: storage,
	}
}

func (svc *odataSpecService) Find(kind, id string) (*storage.ODataSpec, error) {
	key := fmt.Sprintf("%s/%s", kind, id)
	odataSpec, exists, err := svc.storage.ODataSpec(key)
	if !exists || err != nil {
		return nil, err
	}

	return odataSpec, err
}
