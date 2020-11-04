package worker

import (
	"errors"
	"fmt"

	"github.com/kyma-project/kyma/components/binding/internal/storage"
	"github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
)

type TargetKindStorage interface {
	Register(tk v1alpha1.TargetKind) error
	Unregister(tk v1alpha1.TargetKind) error
	Get(kind v1alpha1.Kind) (*storage.ResourceData, error)
	Exist(kind v1alpha1.Kind) bool
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

	if targetKind.Status.Registered {
		if w.storage.Exist(targetKind.Spec.Resource.Kind) {
			log.Infof("TargetKind %q was already registered", targetKind.Name)
			return targetKind, nil
		}
	}
	err := w.storage.Register(*targetKind)
	if err != nil {
		return &v1alpha1.TargetKind{}, errors.New(fmt.Sprintf("while registering TargetKind %q", targetKind.Spec.DisplayName))
	}
	targetKind.Status.Registered = true
	log.Infof("TargetKind %q registered", targetKind.Name)
	return targetKind, nil
}

func (w *TargetKindWorker) RemoveProcess(targetKind *v1alpha1.TargetKind, log log.FieldLogger) error {
	log.Info("start TargetKind removing process")

	return w.storage.Unregister(*targetKind)
}
