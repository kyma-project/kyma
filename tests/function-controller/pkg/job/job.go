package job

import (
	"time"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/resource"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"

	"github.com/pkg/errors"
)

type Job struct {
	resCli             *resource.Resource
	namespace          string
	waitTimeout        time.Duration
	log                shared.Logger
	verbose            bool
	parentFunctionName string
}

func New(parentFunctionName string, c shared.Container) *Job {
	return &Job{
		resCli: resource.New(c.DynamicCli, schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "jobs",
		}, c.Namespace, c.Log, c.Verbose),
		parentFunctionName: parentFunctionName,
		namespace:          c.Namespace,
		waitTimeout:        c.WaitTimeout,
		log:                c.Log,
		verbose:            c.Verbose,
	}
}

func (j Job) List() (*batchv1.JobList, error) {
	selector := map[string]string{
		serverlessv1alpha1.FunctionNameLabel: j.parentFunctionName,
	}

	u, err := j.resCli.List(selector)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing Jobs by label %s in namespace %s", labels.SelectorFromSet(selector).String(), j.namespace)
	}

	jobs, err := convertFromUnstructuredToJobList(u)
	if err != nil {
		return nil, err
	}

	return &jobs, nil
}

func convertFromUnstructuredToJobList(u *unstructured.UnstructuredList) (batchv1.JobList, error) {
	jobList := batchv1.JobList{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), &jobList)
	return jobList, err
}
