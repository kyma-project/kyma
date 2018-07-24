package content

import (
	"fmt"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/storage"
)

type contentService struct {
	storage minioContentGetter
}

func newContentService(storage minioContentGetter) *contentService {
	return &contentService{
		storage: storage,
	}
}

func (svc *contentService) Find(kind, id string) (*storage.Content, error) {
	key := fmt.Sprintf("%s/%s", kind, id)
	content, exists, err := svc.storage.Content(key)
	if !exists || err != nil {
		return nil, err
	}

	return content, err
}
