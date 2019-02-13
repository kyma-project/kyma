package store_test

import (
	"fmt"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/store"
	"github.com/onsi/gomega"
	"testing"
)

func TestBucketName(t *testing.T) {
	//Given
	g := gomega.NewGomegaWithT(t)

	name := "ala"
	namespace := "ola"
	expected := fmt.Sprintf("ns-%s-%s", namespace, name)

	//When
	result := store.BucketName(namespace, name)

	//Then
	g.Expect(result).To(gomega.Equal(expected))
}
