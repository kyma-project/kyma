package mapping

func (c *Controller) WithMappingLister(svc mappingLister) *Controller{
	c.mappingSvc = svc
	return c
}