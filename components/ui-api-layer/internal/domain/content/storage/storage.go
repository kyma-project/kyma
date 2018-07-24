package storage

type Service interface {
	ApiSpec(id string) (*ApiSpec, bool, error)
	AsyncApiSpec(id string) (*AsyncApiSpec, bool, error)
	Content(id string) (*Content, bool, error)
	Initialize(stop <-chan struct{})
}

func New(minio Minio, cache Cache, bucketName, externalAddress, assetsFolder string) Service {
	client := newMinioClient(minio)
	store := newStore(client, bucketName, externalAddress, assetsFolder)
	return newCache(store, cache)
}
