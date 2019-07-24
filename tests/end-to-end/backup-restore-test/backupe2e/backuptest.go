package backupe2e

type BackupTest interface {
	CreateResources(namespace string)
	TestResources(namespace string)
	DeleteResources(namespace string)
}
