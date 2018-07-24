package automock

// EventActivation

func NewEventActivationLister() *eventActivationLister {
	return new(eventActivationLister)
}

func NewReSvc() *reSvc {
	return new(reSvc)
}

func NewStatusGetter() *statusGetter {
	return new(statusGetter)
}
