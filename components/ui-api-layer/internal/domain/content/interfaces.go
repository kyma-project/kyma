package content

import (
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/storage"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
)

//go:generate mockery -name=topicsConverterInterface -output=automock -outpkg=automock -case=underscore
type topicsConverterInterface interface {
	ToGQL(in []gqlschema.TopicEntry) *gqlschema.JSON
	ExtractSection(documents []storage.Document, internal bool) ([]gqlschema.Section, error)
}

//go:generate mockery -name=contentGetter -output=automock -outpkg=automock -case=underscore
type contentGetter interface {
	Find(kind, id string) (*storage.Content, error)
}

//go:generate mockery -name=minioAsyncApiSpecGetter -output=automock -outpkg=automock -case=underscore
type minioAsyncApiSpecGetter interface {
	AsyncApiSpec(id string) (*storage.AsyncApiSpec, bool, error)
}

//go:generate mockery -name=minioApiSpecGetter -output=automock -outpkg=automock -case=underscore
type minioApiSpecGetter interface {
	ApiSpec(id string) (*storage.ApiSpec, bool, error)
}

//go:generate mockery -name=minioContentGetter -output=automock -outpkg=automock -case=underscore
type minioContentGetter interface {
	Content(id string) (*storage.Content, bool, error)
}
