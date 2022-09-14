package compassruntimeagentinit

type RollbackFunc func() error

func newRollbackFunc(runtimeID string, directorClient DirectorClient, rollbackFunctions ...RollbackFunc) RollbackFunc {
	return func() error {
		if err := directorClient.UnregisterRuntime(runtimeID); err != nil {
			return err
		}

		for _, f := range rollbackFunctions {
			if f != nil {
				if err := f(); err != nil {
					return err
				}
			}
		}

		return nil
	}
}
