package serverless

import (
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/kyma-project/kyma/components/function-controller/internal/git"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

func (r *FunctionReconciler) syncRevision(instance *serverlessv1alpha1.Function, options git.Options) (string, error) {
	if instance.Spec.Type == serverlessv1alpha1.SourceTypeGit {
		return r.gitOperator.LastCommit(options)
	}
	return "", nil
}

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
