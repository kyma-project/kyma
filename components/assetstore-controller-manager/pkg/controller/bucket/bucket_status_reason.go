package bucket

type BucketStatusReason string

const (
	ReasonNotFound           BucketStatusReason = "BucketNotFound"
	ReasonPolicyUpdateFailed BucketStatusReason = "BucketPolicyUpdateFailed"
	BucketCreationFailure    BucketStatusReason = "BucketCreationFailure"
	BucketCreated            BucketStatusReason = "BucketCreated"
	BucketPolicyUpdated      BucketStatusReason = "BucketPolicyUpdated"
	BucketPolicyUpdateFailed BucketStatusReason = "BucketPolicyUpdateFailed"
)

func (r BucketStatusReason) String() string {
	return string(r)
}
