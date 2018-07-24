package automock

func NewContentGetter() *contentGetter {
	return new(contentGetter)
}

func NewMinioContentGetter() *minioContentGetter {
	return new(minioContentGetter)
}

func NewMinioAsyncApiSpecGetter() *minioAsyncApiSpecGetter {
	return new(minioAsyncApiSpecGetter)
}

func NewMinioApiSpecGetter() *minioApiSpecGetter {
	return new(minioApiSpecGetter)
}

func NewMockTopicsConverter() *topicsConverterInterface {
	return new(topicsConverterInterface)
}
