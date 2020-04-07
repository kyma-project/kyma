package client

import "testing"

type BackupTest interface {
	CreateResources(t *testing.T, namespace string)
	TestResources(t *testing.T, namespace string)
}
