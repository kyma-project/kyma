package applications

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

//go:generate mockery --name=Comparator
type Comparator interface {
	Compare(test *testing.T, expected, actual string) error
}

func NewSecretComparator(coreClientSet kubernetes.Interface, expectedNamespace, actualNamespace string) (Comparator, error) {
	return &secretComparator{
		coreClientSet:     coreClientSet,
		expectedNamespace: expectedNamespace,
		actualNamespace:   actualNamespace,
	}, nil
}

type secretComparator struct {
	coreClientSet     kubernetes.Interface
	expectedNamespace string
	actualNamespace   string
}

func (c secretComparator) Compare(t *testing.T, expected, actual string) error {
	t.Helper()

	if actual == "" && expected == "" {
		return nil
	}

	if actual == "" || expected == "" {
		return errors.New("empty actual or expected secret name")
	}

	expectedSecretRepo := c.coreClientSet.CoreV1().Secrets(c.expectedNamespace)
	actualSecretRepo := c.coreClientSet.CoreV1().Secrets(c.actualNamespace)

	expectedSecret, err := expectedSecretRepo.Get(context.Background(), expected, metav1.GetOptions{})
	if err != nil {
		return err
	}

	actualSecret, err := actualSecretRepo.Get(context.Background(), actual, metav1.GetOptions{})
	if err != nil {
		return err
	}

	require.Equal(t, expectedSecret.Data, actualSecret.Data)

	return nil
}
