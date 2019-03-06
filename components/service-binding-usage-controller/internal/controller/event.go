package controller

// SBUDeletedEvent contains information about deleted ServiceBindingUsage
type SBUDeletedEvent struct {
	UsedByKind string
	Namespace  string
	Name       string
}
