package controller

func (c *ServiceBindingUsageController) WithTestHookOnAsyncOpDone(h func()) *ServiceBindingUsageController {
	c.testHookAsyncOpDone = h
	return c
}

func (c *ServiceBindingUsageController) WithoutRetries() *ServiceBindingUsageController {
	c.maxRetires = 0
	return c
}
