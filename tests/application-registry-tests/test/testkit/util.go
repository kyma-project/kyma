package testkit

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/rand"
)

func GenerateIdentifier() string {
	return fmt.Sprintf("identifier-%s", rand.String(8))
}
