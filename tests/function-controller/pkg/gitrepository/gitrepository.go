package gitrepository

import (
	"time"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/helpers"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/sirupsen/logrus"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/resource"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GitRepository struct {
	resCli      *resource.Resource
	name        string
	namespace   string
	waitTimeout time.Duration
	log         *logrus.Entry
	verbose     bool
}

func New(name string, c shared.Container) *GitRepository {
	return &GitRepository{
		resCli:      resource.New(c.DynamicCli, serverlessv1alpha1.GroupVersion.WithResource("gitrepositories"), c.Namespace, c.Log, c.Verbose),
		name:        name,
		namespace:   c.Namespace,
		waitTimeout: c.WaitTimeout,
		log:         c.Log,
		verbose:     c.Verbose,
	}
}

func (r GitRepository) GetName() string {
	return r.name
}

func (r *GitRepository) Create(spec serverlessv1alpha1.GitRepositorySpec) error {
	repo := &serverlessv1alpha1.GitRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitRepository",
			APIVersion: serverlessv1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.name,
			Namespace: r.namespace,
		},
		Spec: spec,
	}

	_, err := r.resCli.Create(repo)
	return errors.Wrapf(err, "while creating GitRepository %s in namespace %s", r.name, r.namespace)
}

func (r *GitRepository) Delete() error {
	err := r.resCli.Delete(r.name, r.waitTimeout)

	return errors.Wrapf(err, "while deleting GitRepository %s in namespace %s", r.name, r.namespace)
}

func (r *GitRepository) LogResource() error {
	gitRepo, err := r.Get()
	if err != nil {
		return errors.Wrapf(err, "while getting git repository")
	}

	out, err := helpers.PrettyMarshall(gitRepo)
	if err != nil {
		return errors.Wrap(err, "while marshalling git repository")
	}

	r.log.Infof("GitRepository: %s", out)
	return nil
}

func (r *GitRepository) Get() (serverlessv1alpha1.GitRepository, error) {
	out, err := r.resCli.Get(r.name)
	if err != nil {
		return serverlessv1alpha1.GitRepository{}, errors.Wrap(err, "while getting git repository from cluster")
	}
	repo := serverlessv1alpha1.GitRepository{}

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(out.Object, &repo)
	if err != nil {
		return serverlessv1alpha1.GitRepository{}, errors.Wrapf(err, "while creating object from unstructured object")
	}
	return repo, err
}
