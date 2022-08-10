package application_connectivity_validator

func (gs *ValidatorSuite) TestSimple() {
	gs.Run("Test just the boolean", func() {
		gs.True(true)
	})
}
