package bundle

func (c *RepositoryController) WithTestHookOnAsyncOpDone(h func()) *RepositoryController {
	c.testHookAsyncOpDone = h
	return c
}

func (c *RepositoryController) WithoutRetries() *RepositoryController {
	c.maxRetires = 0
	return c
}
