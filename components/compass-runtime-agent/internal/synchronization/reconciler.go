package synchronization

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
	ApplicationEntry ApplicationEntry
}

func (r Reconciler) Do(applications []ApplicationEntry) (error, []Action) {
	return nil, nil
}
