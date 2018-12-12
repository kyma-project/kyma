package syncer

func (c *Controller) WithCRValidator(validator remoteEnvironmentCRValidator) *Controller {
	c.reCRValidator = validator
	return c
}

func (c *Controller) WithCRMapper(mapper remoteEnvironmentCRMapper) *Controller {
	c.reCRMapper = mapper
	return c
}
