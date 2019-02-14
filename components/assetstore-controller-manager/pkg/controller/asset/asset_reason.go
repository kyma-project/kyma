package asset

type AssetReason string

const (
	ReasonScheduled      AssetReason = "Scheduled"
	ReasonReady          AssetReason = "Ready"
	ReasonPulled         AssetReason = "Pulled"
	ReasonUploaded       AssetReason = "Uploaded"
	ReasonBucketNotReady AssetReason = "BucketNotReady"
	ReasonError          AssetReason = "Error"

	ReasonMutated        AssetReason = "Mutated"
	ReasonMutationFailed AssetReason = "MutationFailed"

	ReasonValidated        AssetReason = "Validated"
	ReasonValidationFailed AssetReason = "ValidationFailed"

	ReasonMissingFiles AssetReason = "MissingFiles"
)

type EventLevel string

const (
	EventWarning EventLevel = "Warning"
	EventNormal  EventLevel = "Normal"
)
