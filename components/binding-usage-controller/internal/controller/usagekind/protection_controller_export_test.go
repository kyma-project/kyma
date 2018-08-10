package usagekind

const FinalizerName = finalizerName

func (c *ProtectionController) WithTestHookOnAsyncOpDone(addHook func(), deletionHook func()) *ProtectionController {
	c.testHookAddFinalizerDone = addHook
	c.testHookProcessDeletionDone = deletionHook
	return c
}
