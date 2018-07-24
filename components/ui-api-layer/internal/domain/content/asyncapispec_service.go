package content

import (
	"fmt"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/storage"
)

type asyncApiSpecService struct {
	storage minioAsyncApiSpecGetter
}

func newAsyncApiSpecService(storage minioAsyncApiSpecGetter) *asyncApiSpecService {
	return &asyncApiSpecService{
		storage: storage,
	}
}

func (svc *asyncApiSpecService) Find(kind, id string) (*storage.AsyncApiSpec, error) {
	key := fmt.Sprintf("%s/%s", kind, id)
	asyncApiSpec, exists, err := svc.storage.AsyncApiSpec(key)
	if !exists || err != nil {
		return nil, err
	}

	return asyncApiSpec, err
}
