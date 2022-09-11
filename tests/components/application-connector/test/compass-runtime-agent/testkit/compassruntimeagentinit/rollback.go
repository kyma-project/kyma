package compassruntimeagentinit

type RollbackFunc func() error

// TODO: Consider changing this interface to the following, as it would be more convenient
// func aggregate(funcs ...RollbackFunc) RollbackFunc

func newRollbackFunc(runtimeID string, directorClient DirectorClient, secretRollback RollbackSecretFunc, deploymentRollback RollbackDeploymentFunc) RollbackFunc {
	return func() error {
		err := directorClient.UnregisterRuntime(runtimeID)
		if err != nil {
			return err
		}

		if secretRollback != nil {
			return secretRollback()
		}

		if deploymentRollback != nil {
			return deploymentRollback()
		}

		return nil
	}
}
