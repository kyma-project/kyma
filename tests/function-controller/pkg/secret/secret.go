package secret

import (
	"time"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/helpers"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/resource"
)

type Secret struct {
	resCli      *resource.Resource
	name        string
	namespace   string
	waitTimeout time.Duration
	log         *logrus.Entry
	verbose     bool
}

func NewSecret(name string, c shared.Container) *Secret {
	return &Secret{
		resCli:      resource.New(c.DynamicCli, corev1.SchemeGroupVersion.WithResource("secrets"), c.Namespace, c.Log, c.Verbose),
		name:        name,
		namespace:   c.Namespace,
		waitTimeout: c.WaitTimeout,
		log:         c.Log,
		verbose:     c.Verbose,
	}
}

func (s *Secret) Name() string {
	return s.name
}

func (s *Secret) Create(data map[string]string) error {
	cm := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.name,
			Namespace: s.namespace,
		},
		StringData: data,
	}

	_, err := s.resCli.Create(cm)
	if err != nil {
		return errors.Wrapf(err, "while creating Secret %s in namespace %s", s.name, s.namespace)
	}
	return err
}

func (s *Secret) Delete() error {
	err := s.resCli.Delete(s.name)
	if err != nil {
		return errors.Wrapf(err, "while deleting Secret %s in namespace %s", s.name, s.namespace)
	}

	return nil
}

func (s *Secret) Get() (*corev1.Secret, error) {
	u, err := s.resCli.Get(s.name)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Secret %s in namespace %s", s.name, s.namespace)
	}
	secret := corev1.Secret{}
	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &secret); err != nil {
		return nil, errors.Wrap(err, "while constructing Secret from unstructured")
	}

	return &secret, nil
}
func (s *Secret) LogResource() error {
	secret, err := s.Get()
	if err != nil {
		return errors.Wrap(err, "while getting Secret")
	}

	out, err := helpers.PrettyMarshall(secret)
	if err != nil {
		return err
	}

	s.log.Infof("Secret: %s", out)
	return nil
}
