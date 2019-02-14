package store

import (
	"fmt"
)

func BucketName(namespace, bucket string) string {
	return fmt.Sprintf("ns-%s-%s", namespace, bucket)
}
