package assethook_test

import (
	"context"
	"fmt"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/assethook"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/assethook/automock"
	"github.com/onsi/gomega"
	"testing"
)

func TestMutationEngine_Mutate(t *testing.T) {
	for testName, testCase := range map[string]struct {
		err      error
		messages map[string][]assethook.Message
	}{
		"success": {},
		"error": {
			err: fmt.Errorf("test"),
		},
		"fail": {
			messages: map[string][]assethook.Message{
				"test": {
					{Filename: "test", Message: "test"},
				},
			},
		},
	} {
		t.Run(testName, func(t *testing.T) {
			// Given
			g := gomega.NewGomegaWithT(t)

			processor := automock.NewHttpProcessor()
			defer processor.AssertExpectations(t)
			ctx := context.TODO()
			files := []string{}
			services := []v1alpha2.AssetWebhookService{}

			processor.On("Do", ctx, "", files, services).Return(testCase.messages, testCase.err).Once()
			mutator := assethook.NewTestMutator(processor)

			// When
			result, err := mutator.Mutate(ctx, "", files, services)

			// Then
			if testCase.err == nil {
				g.Expect(err).ToNot(gomega.HaveOccurred())
			} else {
				g.Expect(err).To(gomega.HaveOccurred())
			}

			if len(testCase.messages) == 0 && err == nil {
				g.Expect(result.Success).To(gomega.BeTrue())
			} else {
				g.Expect(result.Success).To(gomega.BeFalse())
			}
		})
	}
}
