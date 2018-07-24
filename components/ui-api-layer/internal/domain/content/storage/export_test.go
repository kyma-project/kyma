package storage

func NewCache(store storeGetter, cacheClient Cache) *cache {
	return newCache(store, cacheClient)
}

func NewStore(client client, bucketName, externalAddress, assetFolder string) *store {
	return newStore(client, bucketName, externalAddress, assetFolder)
}

func NewNotificationChan() chan notification {
	return make(chan notification)
}

func GetDirectNotificationChan(notifications chan notification) <-chan notification {
	return notifications
}

func NewNotification() notification {
	return notification{}
}

func NewStoreGetter() *mockStoreGetter {
	return new(mockStoreGetter)
}

func NewMockClient() *mockClient {
	return new(mockClient)
}

func NewMinioClient(minio Minio) *minioClient {
	return newMinioClient(minio)
}
