package internal

const (
	BindingInitialization    = "binding initialization"
	BindingReady             = "binding successfully processed"
	BindingTargetFailed      = "binding failed during handling target: %s"
	BindingPendingFromFailed = "binding failed; reprocessing"
	BindingRemovingFailed    = "binding failed during removing process: %s"
	BindingValidationFailed  = "failed validation label %s should be present on the Binding object"
)
