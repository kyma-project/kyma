package pretty

type Reason string

const (
	BucketNotFound                 Reason = "BucketNotFound"
	BucketCreationFailure          Reason = "BucketCreationFailure"
	BucketVerificationFailure      Reason = "BucketVerificationFailure"
	BucketCreated                  Reason = "BucketCreated"
	BucketPolicyUpdated            Reason = "BucketPolicyUpdated"
	BucketPolicyUpdateFailed       Reason = "BucketPolicyUpdateFailed"
	BucketPolicyVerificationFailed Reason = "BucketPolicyVerificationFailed"
	BucketPolicyHasBeenChanged     Reason = "BucketPolicyHasBeenChanged"
)

func (r Reason) String() string {
	return string(r)
}

func (r Reason) Message() string {
	switch r {
	case BucketCreated:
		return "Bucket has been created"
	case BucketNotFound:
		return "Bucket %s doesn't exist anymore"
	case BucketCreationFailure:
		return "Bucket couldn't be created due to error %s"
	case BucketVerificationFailure:
		return "Bucket couldn't be verified due to error %s"
	case BucketPolicyUpdated:
		return "Bucket policy has been updated"
	case BucketPolicyUpdateFailed:
		return "Bucket policy couldn't be set due to error %s"
	case BucketPolicyVerificationFailed:
		return "Bucket policy couldn't be verified due to error %s"
	case BucketPolicyHasBeenChanged:
		return "Remote bucket policy has been changed"
	default:
		return ""
	}
}
