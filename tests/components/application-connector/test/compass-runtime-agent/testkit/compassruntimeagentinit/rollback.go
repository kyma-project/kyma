package compassruntimeagentinit

type RollbackFunc func() error

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
