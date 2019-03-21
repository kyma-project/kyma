package content

func NewContentResolver(contentGetter contentGetter) *contentResolver {
	return newContentResolver(contentGetter)
}

func NewTopicsResolver(contentGetter contentGetter) *topicsResolver {
	return newTopicsResolver(contentGetter)
}

func NewContentService(storage minioContentGetter) *contentService {
	return newContentService(storage)
}

func NewApiSpecService(storage minioApiSpecGetter) *apiSpecService {
	return newApiSpecService(storage)
}

func NewOpenApiSpecService(storage minioOpenApiSpecGetter) *openApiSpecService {
	return newOpenApiSpecService(storage)
}

func NewODataSpecService(storage minioODataSpecGetter) *odataSpecService {
	return newODataSpecService(storage)
}

func NewAsyncApiSpecService(storage minioAsyncApiSpecGetter) *asyncApiSpecService {
	return newAsyncApiSpecService(storage)
}

func (r *topicsResolver) SetTopicsConverter(converter topicsConverterInterface) {
	r.converter = converter
}
