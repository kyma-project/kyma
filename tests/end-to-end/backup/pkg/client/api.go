package client

type BackupTest interface {
	CreateResources(namespace string)
	TestResources(namespace string)
}
