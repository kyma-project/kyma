package syncer

func (c *Controller) WithCRValidator(validator applicationCRValidator) *Controller {
	c.reCRValidator = validator
	return c
}

func (c *Controller) WithCRMapper(mapper applicationCRMapper) *Controller {
	c.reCRMapper = mapper
	return c
}
