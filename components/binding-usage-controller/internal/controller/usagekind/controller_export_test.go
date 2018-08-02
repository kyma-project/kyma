package usagekind

func (c *Controller) WithTestHookOnAsyncOpDone(h func()) *Controller {
	c.testHookAsyncOpDone = h
	return c
}
