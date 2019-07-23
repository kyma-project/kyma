package synchronization

import "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass"

type Reconciler struct {
}

type Operation int

const (
	Create Operation = iota
	Update
	Delete
)

type Action struct {
	Operation        Operation
	ApplicationEntry compass.Application
}

func (r Reconciler) Do(applications []compass.Application) (error, []Action) {
	return nil, nil
}
