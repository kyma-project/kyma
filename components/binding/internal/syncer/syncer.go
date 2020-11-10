package syncer

import (
	"context"
	"github.com/kyma-project/kyma/components/binding/internal/worker"
	bindingsv1alpha1 "github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
)

type Executor struct {
	cli    client.Client
	logger log.FieldLogger
}

func NewExecutor(cli client.Client, logger log.FieldLogger) *Executor {
	return &Executor{
		cli:    cli,
		logger: logger,
	}
}

// TargetKinds assures that existing TargetKinds will be loaded before reconcilers are started
func (e *Executor) TargetKinds(worker *worker.TargetKindWorker) error {
	log.Info("Starting syncing TargetKinds on startup")

	tks := &bindingsv1alpha1.TargetKindList{}
	err := e.cli.List(context.Background(), tks)
	if err != nil {
		return errors.Wrap(err, "while listing TargetKinds")
	}

	tkList := tks.Items
	sort.Slice(tkList, func(i, j int) bool {
		return tkList[i].CreationTimestamp.Before(&tkList[j].CreationTimestamp)
	})

	for _, tk := range tkList {
		_, err := worker.Process(&tk, e.logger.WithField("TargetKind",
			types.NamespacedName{Name: tk.Name, Namespace: tk.Namespace}))
		if err != nil {
			return errors.Wrapf(err, "while processing TargetKind %s/%s", tk.Name, tk.Namespace)
		}
	}

	return nil
}
