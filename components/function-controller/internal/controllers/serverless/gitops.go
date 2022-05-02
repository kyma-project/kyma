package serverless

import (
	"fmt"

	"github.com/kyma-project/kyma/components/function-controller/internal/git"
	ctrl "sigs.k8s.io/controller-runtime"
)

func NextRequeue(err error) (res ctrl.Result, errMsg string) {
	if git.IsNotRecoverableError(err) {
		return ctrl.Result{Requeue: false}, fmt.Sprintf("Stop reconciliation, reason: %s", err.Error())
	}

	errMsg = fmt.Sprintf("Sources update failed, reason: %v", err)
	if git.IsAuthErr(err) {
		errMsg = "Authorization to git server failed"
	}

	// use exponential delay
	return ctrl.Result{Requeue: true}, errMsg
}
