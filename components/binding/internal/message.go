package internal

const (
	BindingInitialization    = "binding initialization"
	BindingReady             = "binding successfully processed"
	BindingTargetFailed      = "binding failed during handling target: %s"
	BindingPendingFromFailed = "binding failed; reprocessing"
	BindingRemovingFailed    = "binding failed during removing process: %s"
	BindingValidationFailed  = "binding validation failed: target or source not exist"
)
