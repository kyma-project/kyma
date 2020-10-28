package worker

import (
	"errors"
	"fmt"

	"github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"
	log "github.com/sirupsen/logrus"
)

type TargetKindStorage interface {
	Register(tk v1alpha1.TargetKind) error
	Unregister(tk v1alpha1.TargetKind) error
	Get(targetKindName string) (v1alpha1.TargetKind, error)
}

type TargetKindWorker struct {
	storage TargetKindStorage
}

func NewTargetKindWorker(storage TargetKindStorage) *TargetKindWorker {
	return &TargetKindWorker{storage: storage}
}

func (w *TargetKindWorker) Process(targetKind *v1alpha1.TargetKind, log log.FieldLogger) (*v1alpha1.TargetKind, error) {
	log.Info("start TargetKind process")

	if targetKind.Spec.Registered {
	_, err := w.storage.Get(targetKind.Name)
		if err != nil {
			err = w.registerTargetKind(*targetKind)
			if err != nil {
				return &v1alpha1.TargetKind{}, errors.New(fmt.Sprintf("while registering TargetKind %q", targetKind.Spec.DisplayName))
			}
			targetKind.Spec.Registered = true
			return targetKind, nil
		}
		return targetKind, nil
	}
	err := w.registerTargetKind(*targetKind)
	if err != nil {
		return &v1alpha1.TargetKind{}, errors.New(fmt.Sprintf("while registering TargetKind %q", targetKind.Spec.DisplayName))
	}
	targetKind.Spec.Registered = true
	log.Infof("TargetKind %q registered", targetKind.Name)
	return targetKind, nil
}

func (w *TargetKindWorker) registerTargetKind(kind v1alpha1.TargetKind) error {
	return w.storage.Register(kind)
}