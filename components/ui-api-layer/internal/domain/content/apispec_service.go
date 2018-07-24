package content

import (
	"fmt"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/storage"
)

type apiSpecService struct {
	storage minioApiSpecGetter
}

func newApiSpecService(storage minioApiSpecGetter) *apiSpecService {
	return &apiSpecService{
		storage: storage,
	}
}

func (svc *apiSpecService) Find(kind, id string) (*storage.ApiSpec, error) {
	key := fmt.Sprintf("%s/%s", kind, id)
	apiSpec, exists, err := svc.storage.ApiSpec(key)
	if !exists || err != nil {
		return nil, err
	}

	return apiSpec, err
}
