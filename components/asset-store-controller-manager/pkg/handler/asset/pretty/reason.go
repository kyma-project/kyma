package pretty

type Reason string

const (
	Pulled                         Reason = "Pulled"
	PullingFailed                  Reason = "PullingFailed"
	Uploaded                       Reason = "Uploaded"
	UploadFailed                   Reason = "UploadFailed"
	BucketNotReady                 Reason = "BucketNotReady"
	BucketError                    Reason = "BucketError"
	Mutated                        Reason = "Mutated"
	MutationFailed                 Reason = "MutationFailed"
	Validated                      Reason = "Validated"
	ValidationFailed               Reason = "ValidationFailed"
	ValidationError                Reason = "ValidationError"
	MissingContent                 Reason = "MissingContent"
	RemoteContentVerificationError Reason = "RemoteContentVerificationError"
	CleanupError                   Reason = "CleanupError"
	Cleaned                        Reason = "Cleaned"
	Scheduled                      Reason = "Scheduled"
)

func (r Reason) String() string {
	return string(r)
}

func (r Reason) Message() string {
	switch r {
	case Pulled:
		return "Asset content has been pulled"
	case PullingFailed:
		return "Asset content pulling failed due to error %s"
	case Uploaded:
		return "Asset content has been uploaded"
	case UploadFailed:
		return "Asset content uploading failed due to error %s"
	case BucketNotReady:
		return "Referenced bucket is not ready"
	case BucketError:
		return "Reading bucket status failed due to error %s"
	case Mutated:
		return "Asset content has been mutated"
	case MutationFailed:
		return "Asset mutation failed due to error %s"
	case Validated:
		return "Asset content has been validated"
	case ValidationFailed:
		return "Asset validation failed due to %+v"
	case ValidationError:
		return "Asset validation failed due to error %s"
	case MissingContent:
		return "Asset content has been removed from remote storage"
	case RemoteContentVerificationError:
		return "Asset content verification failed due to error %s"
	case CleanupError:
		return "Removing old asset content failed due to error %s"
	case Cleaned:
		return "Old asset content hes been removed"
	case Scheduled:
		return "Asset scheduled for processing"
	default:
		return ""
	}
}
