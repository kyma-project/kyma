package finalizer

type finalizer struct {
	name string
}

type FinalizerManager interface {
	GetFinalizers() []string
	SetFinalizers(finalizers []string)
}

type Finalizer interface {
	AddTo(instance FinalizerManager)
	DeleteFrom(instance FinalizerManager)
	IsDefinedIn(instance FinalizerManager) bool
}

func New(name string) Finalizer {
	return &finalizer{
		name: name,
	}
}

func (f *finalizer) AddTo(instance FinalizerManager) {
	finalizers := instance.GetFinalizers()

	if f.containsString(finalizers, f.name) {
		return
	}

	finalizers = append(finalizers, f.name)
	instance.SetFinalizers(finalizers)
}

func (f *finalizer) DeleteFrom(instance FinalizerManager) {
	finalizers := instance.GetFinalizers()
	finalizers = f.removeString(finalizers, f.name)
	instance.SetFinalizers(finalizers)
}

func (f *finalizer) IsDefinedIn(instance FinalizerManager) bool {
	return f.containsString(instance.GetFinalizers(), f.name)
}

func (f *finalizer) containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func (f *finalizer) removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
