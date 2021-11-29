package serverless

import (
	"github.com/kyma-project/kyma/components/function-controller/internal/git"
	"github.com/prometheus/common/log"
)
//TODO: check if backoff doesn't block the go routine
func cloningBackoff(gitClient GitOperator, options git.Options, fnName string, commit *string) func() (done bool, err error) {
	return func() (done bool, err error) {
		id, err := gitClient.LastCommit(options)
		if err != nil {
			log.Warnf("Unable to get last commit from function: %s, cause: %s", fnName, err.Error())
			return false, nil
		}
		commit = &id
		return true, nil
	}
}
