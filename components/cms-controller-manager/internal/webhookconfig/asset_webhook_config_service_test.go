package webhookconfig_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-project/kyma/components/cms-controller-manager/internal/webhookconfig"
	"github.com/kyma-project/kyma/components/cms-controller-manager/internal/webhookconfig/automock"
	"github.com/onsi/gomega"
	_ "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"k8s.io/api/core/v1"
)

var (
	webhookCfgMapName      = "test"
	webhookCfgMapNamespace = "test"
)

func TestAssetWebhookConfigService(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	t.Run("Get", func(t *testing.T) {

		t.Run("nil result", func(t *testing.T) {
			indexer := automock.Indexer{}
			defer indexer.AssertExpectations(t)
			ctx := context.TODO()
			call := indexer.On("GetByKey", mock.AnythingOfType("string"))
			call.Return(nil, false, nil).Once()
			service := webhookconfig.New(&indexer, webhookCfgMapName, webhookCfgMapNamespace)
			actual, err := service.Get(ctx)
			g.Expect(err).To(gomega.BeNil())
			g.Expect(actual).To(gomega.BeEmpty())
		})

		t.Run("ok", func(t *testing.T) {
			indexer := automock.Indexer{}
			defer indexer.AssertExpectations(t)
			ctx := context.TODO()
			call := indexer.On("GetByKey", mock.AnythingOfType("string"))
			call.Return(mockConfigMap(map[string]string{
				"markdown": `{"validations":[{"name":"markdown-validation"}],"mutations":[{"name":"markdown-mutation"}]}`,
				"openapi":  `{"validations":[{"name":"openapi-validation"}],"mutations":[{"name":"openapi-mutation"}]}`,
				"unknow":   `{"test": ["me"]}`,
			}), true, nil).Once()
			service := webhookconfig.New(&indexer, webhookCfgMapName, webhookCfgMapNamespace)
			_, err := service.Get(ctx)
			g.Expect(err).To(gomega.BeNil())
		})

		t.Run("err", func(t *testing.T) {
			indexer := automock.Indexer{}
			defer indexer.AssertExpectations(t)
			ctx := context.TODO()
			call := indexer.On("GetByKey", mock.AnythingOfType("string"))
			testError := errors.New("test_error")
			call.Return(nil, false, testError).Once()
			service := webhookconfig.New(&indexer, webhookCfgMapName, webhookCfgMapNamespace)
			actual, err := service.Get(ctx)
			g.Expect(err).NotTo(gomega.BeNil())
			g.Expect(actual).To(gomega.BeNil())
		})

		t.Run("err-unmarshal", func(t *testing.T) {
			indexer := automock.Indexer{}
			defer indexer.AssertExpectations(t)
			ctx := context.TODO()
			call := indexer.On("GetByKey", mock.AnythingOfType("string"))
			call.Return(mockConfigMap(map[string]string{
				"openapi": `{"validations":[{"name":"openapi-validation"}],"mutations":[{"name":"openapi-mutation"}]}`,
				"watch":   "me explode",
			}), true, nil).Once()
			service := webhookconfig.New(&indexer, webhookCfgMapName, webhookCfgMapNamespace)
			actual, err := service.Get(ctx)
			g.Expect(err).NotTo(gomega.BeNil())
			g.Expect(actual).To(gomega.BeNil())
		})
	})
}

func mockConfigMap(cfgMapContent map[string]string) *v1.ConfigMap {
	return &v1.ConfigMap{
		Data: cfgMapContent,
	}
}
