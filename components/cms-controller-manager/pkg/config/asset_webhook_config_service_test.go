package config_test

import (
	"context"
	"errors"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/config"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/config/automock"
	"github.com/onsi/gomega"
	_ "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"k8s.io/api/core/v1"
	"testing"
)

func Test_assetWhsConfigService(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	t.Run("Get", func(t *testing.T) {
		client := automock.Client{}
		ctx := context.TODO()
		call := client.On("Get", ctx, mock.Anything, &v1.ConfigMap{})

		t.Run("nil result", func(t *testing.T) {
			call.Return(nil).Once()
			service := config.NewAssetWebHookService(&client)
			actual, err := service.Get(ctx, "", "")
			g.Expect(err).To(gomega.BeNil())
			g.Expect(actual).To(gomega.BeEmpty())
		})

		t.Run("ok", func(t *testing.T) {
			name := "markdown"
			call.Run(mockValidConfigMap()).Return(nil).Once()
			service := config.NewAssetWebHookService(&client)
			actual, err := service.Get(ctx, "", name)
			g.Expect(err).To(gomega.BeNil())
			g.Expect(actual).To(gomega.Equal(actual))
		})

		t.Run("err", func(t *testing.T) {
			name := "markdown"
			testError := errors.New("test_error")
			call.Return(testError).Once()
			service := config.NewAssetWebHookService(&client)
			actual, err := service.Get(ctx, "", name)
			g.Expect(err).NotTo(gomega.BeNil())
			g.Expect(actual).To(gomega.BeNil())
		})

		t.Run("err-unmarshal", func(t *testing.T) {
			name := "openapi"
			call.Return(nil).Run(mockInvalidConfigMap()).Once()
			service := config.NewAssetWebHookService(&client)
			actual, err := service.Get(ctx, "", name)
			g.Expect(err).NotTo(gomega.BeNil())
			g.Expect(actual).To(gomega.BeNil())
		})
	})
}

func mockConfigMap(cfgMapContent map[string]string) func(mock.Arguments) {
	return func(args mock.Arguments) {
		cm := args[2].(*v1.ConfigMap)
		cm.Data = cfgMapContent
	}
}

func mockValidConfigMap() func(args mock.Arguments) {
	return mockConfigMap(map[string]string{
		"markdown": `{"validations":[{"name":"markdown-validation"}],"mutations":[{"name":"markdown-mutation"}]}`,
		"openapi":  `{"validations":[{"name":"openapi-validation"}],"mutations":[{"name":"openapi-mutation"}]}`,
		"unknow":   `{"test": ["me"]}`,
	})
}

func mockInvalidConfigMap() func(args mock.Arguments) {
	return mockConfigMap(map[string]string{
		"openapi": `{"validations":[{"name":"openapi-validation"}],"mutations":[{"name":"openapi-mutation"}]}`,
		"watch":   "me explode",
	})
}
