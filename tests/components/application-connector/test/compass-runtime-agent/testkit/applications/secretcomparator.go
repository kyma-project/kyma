package applications

import (
	"context"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Comparator interface {
	Compare(actual, expected string) error
}

func NewSecretComparator(assertions *require.Assertions, cli kubernetes.Interface, expectedNamespace, actualNamespace string) (Comparator, error) {
	return &secretComparator{
		assertions:        assertions,
		cli:               cli,
		expectedNamespace: expectedNamespace,
		actualNamespace:   actualNamespace,
	}, nil
}

type secretComparator struct {
	assertions        *require.Assertions
	cli               kubernetes.Interface
	expectedNamespace string
	actualNamespace   string
}

func (c secretComparator) Compare(actual, expected string) error {

	expectedSecretRepo := c.cli.CoreV1().Secrets(c.expectedNamespace)
	actualSecretRepo := c.cli.CoreV1().Secrets(c.actualNamespace)

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
