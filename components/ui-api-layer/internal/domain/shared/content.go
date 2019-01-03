package shared

import "github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/storage"

//go:generate mockery -name=ContentRetriever -output=automock -outpkg=automock -case=underscore
type ContentRetriever interface {
	Content() ContentGetter
	ApiSpec() ApiSpecGetter
	AsyncApiSpec() AsyncApiSpecGetter
}

//go:generate mockery -name=AsyncApiSpecGetter -output=automock -outpkg=automock -case=underscore
type AsyncApiSpecGetter interface {
	Find(kind, id string) (*storage.AsyncApiSpec, error)
}

//go:generate mockery -name=ApiSpecGetter -output=automock -outpkg=automock -case=underscore
type ApiSpecGetter interface {
	Find(kind, id string) (*storage.ApiSpec, error)
}

//go:generate mockery -name=ContentGetter -output=automock -outpkg=automock -case=underscore
type ContentGetter interface {
	Find(kind, id string) (*storage.Content, error)
}
