package worker

import (
	bindErr "github.com/kyma-project/kyma/components/binding/internal/errors"
	"github.com/kyma-project/kyma/components/binding/internal/storage"
	"github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
)

type TargetKindStorage interface {
	Register(tk v1alpha1.TargetKind) error
	Unregister(tk v1alpha1.TargetKind) error
	Get(kind v1alpha1.Kind) (*storage.ResourceData, error)
	Exist(kind v1alpha1.TargetKind) bool
	Equal(tk v1alpha1.TargetKind, registeredTk *storage.ResourceData) bool
}

type TargetKindWorker struct {
	storage       TargetKindStorage
	dynamicClient dynamic.Interface
}

func NewTargetKindWorker(storage TargetKindStorage, dynamicClient dynamic.Interface) *TargetKindWorker {
	return &TargetKindWorker{
		storage:       storage,
		dynamicClient: dynamicClient,
	}
}

func (w *TargetKindWorker) Process(targetKind *v1alpha1.TargetKind, log log.FieldLogger) (*v1alpha1.TargetKind, error) {
	log.Info("start TargetKind process")

	registered, err := w.storage.Get(targetKind.Spec.Resource.Kind)
	switch {
	case bindErr.IsNotFound(err):
	case err == nil:
		if targetKind.Status.IsRegistered() {
			log.Infof("TargetKind %s was already registered", targetKind.Name)
			hasChanged, err := w.isDifferentThanRegistered(targetKind, registered)
			if err != nil {
				return targetKind, errors.Wrap(err, "while comparing processed TargetKind with registered one")
			}
			if !hasChanged {
				log.Infof("TargetKind %s is not different than existing one", targetKind.Name)
				return targetKind, nil
			}
			targetKind.Status.Message = "Failed because a TargetKind with the same kind and different properties exists"
			err = w.register(targetKind, v1alpha1.TargetKindFailed, log)
			if err != nil {
				return targetKind, err
			}
		}
		if !targetKind.Status.IsEmpty() {
			return targetKind, nil
		}
	default:
		return targetKind, errors.Wrap(err, "while getting target kind")
	}

	err = w.register(targetKind, v1alpha1.TargetKindRegistered, log)
	if err != nil {
		return targetKind, err
	}

	return targetKind, nil
}

func (w *TargetKindWorker) RemoveProcess(targetKind *v1alpha1.TargetKind, log log.FieldLogger) error {
	log.Info("start TargetKind removing process")

	return w.storage.Unregister(*targetKind)
}

func (w *TargetKindWorker) register(targetKind *v1alpha1.TargetKind, status string, log log.FieldLogger) error {
	err := w.storage.Register(*targetKind)
	if err != nil {
		return errors.Wrapf(err, "while registering TargetKind %q", targetKind.Name)
	}
	switch status {
	case v1alpha1.TargetKindRegistered:
		err = targetKind.Status.Registered()
		if err != nil {
			return errors.Wrapf(err, "while set TargetKind phase to %s", status)
		}
	case v1alpha1.TargetKindFailed:
		err = targetKind.Status.Failed()
		if err != nil {
			return errors.Wrapf(err, "while set TargetKind phase to %s", status)
		}
	}
	log.Info(status)
	return nil
}

func (w *TargetKindWorker) isDifferentThanRegistered(targetKind *v1alpha1.TargetKind, registered *storage.ResourceData) (bool, error) {
	registered, err := w.storage.Get(targetKind.Spec.Resource.Kind)
	if err != nil {
		return false, errors.Wrap(err, "while getting ResourceData")
	}
	if !w.storage.Equal(*targetKind, registered) {
		return true, nil
	}
	return false, nil
}
