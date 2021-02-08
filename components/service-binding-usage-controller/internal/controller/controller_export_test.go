package controller

func (c *ServiceBindingUsageController) WithTestHookOnAsyncOpDone(h func()) *ServiceBindingUsageController {
	c.testHookAsyncOpDone = h
	return c
}
