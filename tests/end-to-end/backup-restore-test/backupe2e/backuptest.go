package backupe2e

var backupTests []BackupTest

type BackupTest interface {
	CreateResources(namespace string)
	TestResources(namespace string)
}

func Register(backupTest BackupTest) {
	backupTests = append(backupTests, backupTest)
}

func Tests() []BackupTest {
	return backupTests
}
