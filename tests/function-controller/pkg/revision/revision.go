package revision

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"knative.dev/serving/pkg/apis/serving"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/resource"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
)

type Revision struct {
	resCli            *resource.Resource
	parentServiceName string
	namespace         string
	waitTimeout       time.Duration
	log               shared.Logger
	verbose           bool
}

func New(parentServiceName string, c shared.Container) *Revision {
	return &Revision{
		resCli:            resource.New(c.DynamicCli, servingv1.SchemeGroupVersion.WithResource("revisions"), c.Namespace, c.Log, c.Verbose),
		parentServiceName: parentServiceName,
		namespace:         c.Namespace,
		waitTimeout:       c.WaitTimeout,
		log:               c.Log,
		verbose:           c.Verbose,
	}
}

func (r *Revision) list() (*servingv1.RevisionList, error) {
	selector := map[string]string{
		"serving.knative.dev/service": r.parentServiceName,
	}

	u, err := r.resCli.List(selector)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing Revisions by label %s in namespace %s", labels.SelectorFromSet(selector).String(), r.namespace)
	}

	revisions, err := convertFromUnstructuredToRevisionList(u)
	if err != nil {
		return nil, err
	}

	return &revisions, nil
}

func (r *Revision) WaitForRevisionCleanup() error {
	revList, err := r.list()
	if err != nil {
		return err
	}

	r.logExistingRevisions(*revList)
	if len(revList.Items) == 1 {
		return nil
	}

	condition := r.areRevisionCleanedUp()

	done := make(chan struct{})

	go func() {
		time.Sleep(5 * time.Minute)
		close(done)
	}()

	return wait.PollUntil(5*time.Second, condition, done)
}

func (r *Revision) areRevisionCleanedUp() wait.ConditionFunc {
	return func() (bool, error) {
		revList, err := r.list()
		if err != nil {
			return false, err
		}

		r.logExistingRevisions(*revList)

		return len(revList.Items) == 1, nil
	}
}

func (r Revision) logExistingRevisions(revs servingv1.RevisionList) {
	if len(revs.Items) != 1 {
		r.log.Logf("There are still %d revisions, need only 1", len(revs.Items))
	} else {
		r.log.Logf("Old revisions are deleted, only 1 left")
	}
	if r.verbose {
		r.log.Logf("%+v", revs)
	}
}

func (r *Revision) VerifyConfigurationGeneration(generation int) error {
	revList, err := r.list()
	if err != nil {
		return err
	}

	cfgGeneration, ok := revList.Items[0].Labels[serving.ConfigurationGenerationLabelKey]

	if !ok {
		r.log.Logf("Revisions: %+v", revList)
		return fmt.Errorf("remaining revision doesn't have %s label", serving.ConfigurationGenerationLabelKey)
	}

	if cfgGeneration != strconv.Itoa(generation) {
		return fmt.Errorf("configurationGeneration label has incorrect value of %s, expecting %d", cfgGeneration, generation)
	}
	return nil
}

func convertFromUnstructuredToRevisionList(u *unstructured.UnstructuredList) (servingv1.RevisionList, error) {
	revisionList := servingv1.RevisionList{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), &revisionList)
	return revisionList, err
}
