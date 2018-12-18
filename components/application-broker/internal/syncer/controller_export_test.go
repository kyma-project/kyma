package syncer

func (c *Controller) WithCRValidator(validator applicationCRValidator) *Controller {
	c.appCRValidator = validator
	return c
}

func (c *Controller) WithCRMapper(mapper applicationCRMapper) *Controller {
	c.appCRMapper = mapper
	return c
}
