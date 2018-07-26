package automock

func NewAzureBlobClient() *azureBlobClient {
	return new(azureBlobClient)
}
