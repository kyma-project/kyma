package automock

// EventActivation

func NewEventActivationLister() *eventActivationLister {
	return new(eventActivationLister)
}

func NewApplicationSvc() *appSvc {
	return new(appSvc)
}

func NewStatusGetter() *statusGetter {
	return new(statusGetter)
}
