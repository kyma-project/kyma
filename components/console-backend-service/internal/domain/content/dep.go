package content

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/content/storage"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

//go:generate mockery -name=topicsConverterInterface -output=automock -outpkg=automock -case=underscore
type topicsConverterInterface interface {
	ToGQL(in []gqlschema.TopicEntry) *gqlschema.JSON
	ExtractSection(documents []storage.Document, internal bool) ([]gqlschema.Section, error)
}

//go:generate mockery -name=minioAsyncApiSpecGetter -output=automock -outpkg=automock -case=underscore
type minioAsyncApiSpecGetter interface {
	AsyncApiSpec(id string) (*storage.AsyncApiSpec, bool, error)
}

//go:generate mockery -name=minioApiSpecGetter -output=automock -outpkg=automock -case=underscore
type minioApiSpecGetter interface {
	ApiSpec(id string) (*storage.ApiSpec, bool, error)
}

//go:generate mockery -name=minioOpenApiSpecGetter -output=automock -outpkg=automock -case=underscore
type minioOpenApiSpecGetter interface {
	OpenApiSpec(id string) (*storage.OpenApiSpec, bool, error)
}

//go:generate mockery -name=minioODataSpecGetter -output=automock -outpkg=automock -case=underscore
type minioODataSpecGetter interface {
	ODataSpec(id string) (*storage.ODataSpec, bool, error)
}

//go:generate mockery -name=minioContentGetter -output=automock -outpkg=automock -case=underscore
type minioContentGetter interface {
	Content(id string) (*storage.Content, bool, error)
}
