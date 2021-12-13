package gitrepository

import (
	"context"
	"fmt"
	"github.com/kyma-project/kyma/components/function-controller/internal/resource"
	serverlessv1alhpa1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type GitRepoController struct {
	client resource.Client
}

func NewGitRepoController(client resource.Client) *GitRepoController {
	return &GitRepoController{client: client}
}
func (c *GitRepoController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).Named("git repository controller").
		For(&serverlessv1alhpa1.GitRepository{}).
		Complete(c)
}

func (c *GitRepoController) Reconcile(request ctrl.Request) (ctrl.Result, error) {
	fmt.Println(" Iam live, ", request.NamespacedName)
	ctx := context.TODO()
	instance := &serverlessv1alpha1.GitRepository{}
	err := c.client.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	instance.Status = serverlessv1alpha1.GitRepositoryStatus{
		LastTransitionTime: metav1.Now(),
		LastCommit:         "Ugabuga",
		Reason:             "Update",
		Message:            "brrrr",
	}

	err = c.client.Status().Update(ctx, instance)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "while updating status")
	}

	return ctrl.Result{
		RequeueAfter: 5 * time.Second,
	}, nil
}
