package automock

func NewConfigMapClient() *configMapClient {
	return new(configMapClient)
}

func NewSingleBackupExecutor() *singleBackupExecutor {
	return new(singleBackupExecutor)
}
