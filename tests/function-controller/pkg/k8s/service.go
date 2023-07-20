package k8s

import (
	"context"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/helpers"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
)

type Service struct {
	name      string
	namespace string
	image     string
	port      int32
	coreCli   coreclient.ServiceInterface
	log       *logrus.Entry
}

func NewService(name, namespace string, port int32, coreCli coreclient.ServiceInterface, log *logrus.Entry) Service {
	return Service{
		name:      name,
		namespace: namespace,
		port:      port,
		coreCli:   coreCli,
		log:       log,
	}
}

func (s Service) Create() error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: s.name,
		},
		Spec: corev1.ServiceSpec{
			Type: "ClusterIP",
			Ports: []corev1.ServicePort{
				{
					Name:       s.name,
					Port:       s.port,
					Protocol:   "TCP",
					TargetPort: intstr.FromInt(int(s.port)),
				},
			},
			Selector: map[string]string{
				componentLabel: s.name,
			},
		},
	}

	_, err := s.coreCli.Create(context.Background(), service, metav1.CreateOptions{})
	return errors.Wrapf(err, "while creating service %s in namespace %s", s.name, s.namespace)
}

func (s Service) Delete(ctx context.Context, options metav1.DeleteOptions) error {
	return s.coreCli.Delete(ctx, s.name, options)
}

func (s Service) Get(ctx context.Context, options metav1.GetOptions) (*corev1.Service, error) {
	svc, err := s.coreCli.Get(ctx, s.name, options)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting service %s in namespace %s", s.name, s.namespace)
	}
	return svc, nil
}
func (s Service) LogResource() error {
	svc, err := s.Get(context.TODO(), metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "while getting service")
	}
	out, err := helpers.PrettyMarshall(svc)
	if err != nil {
		return errors.Wrap(err, "while marshalling service")
	}
	s.log.Info(out)
	return nil
}
