package worker

import (
	"fmt"

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
	return &TargetKindWorker{storage: storage,
		dynamicClient: dynamicClient,
	}
}

func (w *TargetKindWorker) Process(targetKind *v1alpha1.TargetKind, log log.FieldLogger) (*v1alpha1.TargetKind, error) {
	log.Info("start TargetKind process")

	if targetKind.Status.IsRegistered() {
		log.Infof("TargetKind %s was already registered", targetKind.Name)
		if !w.storage.Exist(*targetKind) {
			err := w.storage.Register(*targetKind)
			if err != nil {
				return targetKind, fmt.Errorf("while registering TargetKind %q", targetKind.Name)
			}
			err = targetKind.Status.Registered()
			if err != nil {
				return targetKind, errors.Wrapf(err, "while set TargetKind phase to %s", v1alpha1.TargetKindRegistered)
			}
			return targetKind, nil
		}

		hasChanged, err := w.isDifferentThanRegistered(targetKind)
		if err != nil {
			return targetKind, errors.Wrap(err, "while comparing processed TargetKind with registered one")
		}
		if !hasChanged {
			log.Infof("TargetKind %s is not different than existing one", targetKind.Name)
			return targetKind, nil
		}

		err = w.storage.Register(*targetKind)
		if err != nil {
			return targetKind, fmt.Errorf("while registering TargetKind %q", targetKind.Name)
		}
		err = targetKind.Status.Registered()
		if err != nil {
			return targetKind, errors.Wrapf(err, "while set TargetKind phase to %s", v1alpha1.TargetKindRegistered)
		}

		//TODO: mark old existing TargetKind CR as 'invalid'
		return targetKind, nil

	}
	// TargetKind was not registered before
	err := w.storage.Register(*targetKind)
	if err != nil {
		return targetKind, errors.New(fmt.Sprintf("while registering TargetKind %q", targetKind.Name))
	}
	err = targetKind.Status.Registered()
	if err != nil {
		return targetKind, errors.Wrapf(err, "while set TargetKind phase to %s", v1alpha1.TargetKindRegistered)
	}
	log.Infof("TargetKind %q registered", targetKind.Name)
	return targetKind, nil
}

func (w *TargetKindWorker) RemoveProcess(targetKind *v1alpha1.TargetKind, log log.FieldLogger) error {
	log.Info("start TargetKind removing process")

	return w.storage.Unregister(*targetKind)
}

func (w *TargetKindWorker) isDifferentThanRegistered(targetKind *v1alpha1.TargetKind) (bool, error) {
	registered, err := w.storage.Get(targetKind.Spec.Resource.Kind)
	if err != nil {
		return false, errors.Wrap(err, "while getting ResourceData")
	}
	if !w.storage.Equal(*targetKind, registered) {
		return true, nil
	}
	return false, nil
}
