package content

import (
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/content/storage"
)

type openApiSpecService struct {
	storage minioOpenApiSpecGetter
}

func newOpenApiSpecService(storage minioOpenApiSpecGetter) *openApiSpecService {
	return &openApiSpecService{
		storage: storage,
	}
}

func (svc *openApiSpecService) Find(kind, id string) (*storage.OpenApiSpec, error) {
	key := fmt.Sprintf("%s/%s", kind, id)
	openApiSpec, exists, err := svc.storage.OpenApiSpec(key)
	if !exists || err != nil {
		return nil, err
	}

	return openApiSpec, err
}
