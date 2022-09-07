package applications

import (
	"context"
	"errors"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

//go:generate mockery --name=Comparator
type Comparator interface {
	Compare(expected, actual string) error
}

func NewSecretComparator(assertions *require.Assertions, coreClientSet kubernetes.Interface, expectedNamespace, actualNamespace string) (Comparator, error) {
	return &secretComparator{
		assertions:        assertions,
		coreClientSet:     coreClientSet,
		expectedNamespace: expectedNamespace,
		actualNamespace:   actualNamespace,
	}, nil
}

type secretComparator struct {
	assertions        *require.Assertions
	coreClientSet     kubernetes.Interface
	expectedNamespace string
	actualNamespace   string
}

func (c secretComparator) Compare(expected, actual string) error {

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
	c.assertions.Equal(expectedSecret.Data, actualSecret.Data)

	return nil
}
