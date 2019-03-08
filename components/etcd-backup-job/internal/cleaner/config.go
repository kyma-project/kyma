package cleaner

import "time"

// Config holds AzureCleaner configuration
type Config struct {
	// LeaveMinNewestBackupBlobs defines the number of blobs which should not be deleted even
	// if they are treated as expired.
	LeaveMinNewestBackupBlobs int64

	// ExpirationBlobTime defines delta which is used to check if blob should be deleted
	// if blob.LastModified < (timeNow - expirationBlobTime)  then blob will be removed
	ExpirationBlobTime time.Duration
}
