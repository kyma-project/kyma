package worker

import (
	"fmt"

	"github.com/kyma-project/kyma/components/binding/internal"
	"github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type KindManager interface {
	AddLabel(*v1alpha1.Binding) error
	RemoveLabel(*v1alpha1.Binding) error
	LabelExist(*v1alpha1.Binding) (bool, error)
	RemoveOldAddNewLabel(*v1alpha1.Binding) error
}

type BindingWorker struct {
	kindManager KindManager
}

func NewBindingWorker(km KindManager) *BindingWorker {
	return &BindingWorker{
		kindManager: km,
	}
}

func (b *BindingWorker) Process(binding *v1alpha1.Binding, log log.FieldLogger) (*v1alpha1.Binding, error) {
	log.Info("start Binding process")

	if binding.Status.IsEmpty() {
		log.Info("binding status is empty. Binding initialization.")
		err := binding.Status.Init()
		if err != nil {
			return binding, errors.Wrap(err, "while init Binding phase")
		}
		binding.Status.Source = fmt.Sprintf("%s/%s", binding.Spec.Source.Kind, binding.Spec.Source.Name)
		binding.Status.Message = internal.BindingInitialization
		return binding, nil
	}

	switch binding.Status.Phase {
	case v1alpha1.BindingPending:
		return b.pendingPhase(binding, log)
	case v1alpha1.BindingReady:
		return b.readyPhase(binding, log)
	case v1alpha1.BindingFailed:
		return b.failedPhase(binding, log)
	}

	return binding, errors.Errorf("status phase %s not supported", binding.Status.Phase)
}

// pendingPhase adds label to Target; marks Binding as Ready
func (b *BindingWorker) pendingPhase(binding *v1alpha1.Binding, log log.FieldLogger) (*v1alpha1.Binding, error) {
	log.Infof("set labels to the target: %s - %s", binding.Spec.Target.Kind, binding.Spec.Target.Name)
	err := b.kindManager.AddLabel(binding)
	if err != nil {
		errStatus := binding.Status.Failed()
		if errStatus != nil {
			return binding, errors.Wrapf(errStatus, "while set Binding phase to %s", v1alpha1.BindingFailed)
		}
		binding.Status.Message = fmt.Sprintf(internal.BindingTargetFailed, err)
		return binding, errors.Wrapf(err, "while adding label to target (phase: %s)", v1alpha1.BindingPending)
	}

	log.Info("Binding process successfully completed")
	err = binding.Status.Ready()
	if err != nil {
		return binding, errors.Wrapf(err, "while set Binding phase to %s", v1alpha1.BindingReady)
	}
	binding.Status.Message = internal.BindingReady

	return binding, nil
}

// readyPhase checks if Source was changed; if yes removes old label from Target and adds new
// if Source was not changed checks if label in Target exist, if not adds label to Target
func (b *BindingWorker) readyPhase(binding *v1alpha1.Binding, log log.FieldLogger) (*v1alpha1.Binding, error) {
	if binding.Status.Source != fmt.Sprintf("%s/%s", binding.Spec.Source.Kind, binding.Spec.Source.Name) {
		log.Info("source was changed, removing old label and add new")
		err := b.kindManager.RemoveOldAddNewLabel(binding)
		if err != nil {
			return binding, errors.Wrap(err, "while removing old label and adding new in target")
		}
	}

	labelExist, err := b.kindManager.LabelExist(binding)
	if err != nil {
		return binding, errors.Wrap(err, "while checking if label exist in target")
	}
	if !labelExist {
		log.Info("Binding in %s state but label not exist in target, adding new label", v1alpha1.BindingReady)
		err := b.kindManager.AddLabel(binding)
		if err != nil {
			return binding, errors.Wrapf(err, "while adding label to target (phase: %s)", v1alpha1.BindingReady)
		}
	}

	return binding, nil
}

// failedPhase check if label on target exist; if yes removes old label, adds new and marks Binding as Ready
// if not triggers pending process for Binding
func (b *BindingWorker) failedPhase(binding *v1alpha1.Binding, log log.FieldLogger) (*v1alpha1.Binding, error) {
	labelExist, err := b.kindManager.LabelExist(binding)
	if err != nil {
		return binding, errors.Wrap(err, "while checking if label exist in target")
	}

	if labelExist {
		log.Infof("Binding with phase %s has label on target, removing label", v1alpha1.BindingFailed)
		err := b.kindManager.RemoveLabel(binding)
		if err != nil {
			return binding, errors.Wrap(err, "while removing old label and adding new in target")
		}
	}

	err = binding.Status.Pending()
	if err != nil {
		return binding, errors.Wrapf(err, "while set Binding phase to %s", v1alpha1.BindingPending)
	}
	binding.Status.Message = internal.BindingPendingFromFailed

	return binding, nil
}
