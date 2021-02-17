package job

import (
	"time"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	batchv1typed "k8s.io/client-go/kubernetes/typed/batch/v1"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"

	"github.com/pkg/errors"
)

type Job struct {
	client             batchv1typed.JobInterface
	namespace          string
	waitTimeout        time.Duration
	log                *logrus.Logger
	verbose            bool
	parentFunctionName string
}

func New(parentFunctionName string, batchCli batchv1typed.BatchV1Interface, c shared.Container) *Job {
	return &Job{
		client:             batchCli.Jobs(c.Namespace),
		parentFunctionName: parentFunctionName,
		namespace:          c.Namespace,
		waitTimeout:        c.WaitTimeout,
		verbose:            c.Verbose,
	}
}

func (j Job) List() (*batchv1.JobList, error) {
	labelSelector := labels.SelectorFromSet(map[string]string{serverlessv1alpha1.FunctionNameLabel: j.parentFunctionName}).String()

	jobList, err := j.client.List(metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "while listing Jobs by label %s in namespace %s", labelSelector, j.namespace)
	}
	return jobList, nil
}
