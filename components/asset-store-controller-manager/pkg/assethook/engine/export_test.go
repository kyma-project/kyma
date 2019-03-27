package engine

import (
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/assethook"
	"os"
	"time"
)

func NewTestMutator(webhook assethook.Webhook, timeout time.Duration, fileReader func(filename string) ([]byte, error), fileWriter func(filename string, data []byte, perm os.FileMode) error) Mutator {
	return &mutationEngine{
		webhook:    webhook,
		timeout:    time.Minute,
		fileReader: fileReader,
		fileWriter: fileWriter,
	}
}

func NewTestValidator(webhook assethook.Webhook, timeout time.Duration, fileReader func(filename string) ([]byte, error)) Validator {
	return &validationEngine{
		webhook:    webhook,
		timeout:    time.Minute,
		fileReader: fileReader,
	}
}
