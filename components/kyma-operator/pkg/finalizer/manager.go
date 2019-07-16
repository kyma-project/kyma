package finalizer

// Finalizable exposes function to modify finalizers in kubernetes objects
type Finalizable interface {
	GetFinalizers() []string
	SetFinalizers([]string)
}

// Manager provides functions to modify finalizers in kubernetes objects
type Manager struct {
	finalizerName string
}

// NewManager returns an instance of Manager configured with given finalizerName
func NewManager(finalizerName string) *Manager {
	return &Manager{
		finalizerName: finalizerName,
	}
}

// AddFinalizer adds finalizer to the object
func (m *Manager) AddFinalizer(object Finalizable) {

	finalizers := object.GetFinalizers()

	if !hasFinalizer(finalizers, m.finalizerName) {
		object.SetFinalizers(append(finalizers, m.finalizerName))
	}
}

// RemoveFinalizer removes finalizer from the object
func (m *Manager) RemoveFinalizer(object Finalizable) {

	finalizers := object.GetFinalizers()
	object.SetFinalizers(removeFinalizer(finalizers, m.finalizerName))
}

// HasFinalizer returns true if the object has the finalizer specified in Manager
func (m *Manager) HasFinalizer(object Finalizable) bool {
	return hasFinalizer(object.GetFinalizers(), m.finalizerName)
}

func hasFinalizer(finalizers []string, element string) bool {

	if finalizers == nil || len(finalizers) == 0 {
		return false
	}

	for _, f := range finalizers {
		if f == element {
			return true
		}
	}

	return false
}

func removeFinalizer(finalizers []string, element string) []string {

	if finalizers == nil || len(finalizers) == 0 {
		return finalizers
	}

	var result = []string{}

	for _, f := range finalizers {
		if f != element {
			result = append(result, f)
		}
	}
	return result
}
