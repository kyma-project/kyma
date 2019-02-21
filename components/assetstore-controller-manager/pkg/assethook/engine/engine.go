package engine

import (
	"encoding/json"
	"fmt"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	"path/filepath"
)

type baseEngine struct{}

//go:generate mockery -name=Accessor -output=automock -outpkg=automock -case=underscore
type Accessor interface {
	GetNamespace() string
	GetName() string
}

func (*baseEngine) parseMetadata(metadata *runtime.RawExtension) *json.RawMessage {
	if nil == metadata {
		return nil
	}

	result := json.RawMessage(metadata.Raw)
	return &result
}

func (*baseEngine) readFiles(basePath string, files []string, reader func(filename string) ([]byte, error)) (map[string]string, error) {
	result := make(map[string]string)

	for _, f := range files {
		path := filepath.Join(basePath, f)
		data, err := reader(path)
		if err != nil {
			return nil, errors.Wrapf(err, "while reading file %s", path)
		}

		result[f] = string(data)
	}

	return result, nil
}

func (*baseEngine) writeFiles(basePath string, content map[string]string, writer func(filename string, data []byte, perm os.FileMode) error) error {
	for key, value := range content {
		path := filepath.Join(basePath, key)
		err := writer(path, []byte(value), os.ModePerm)
		if err != nil {
			return errors.Wrapf(err, "while writing file %s", path)
		}
	}

	return nil
}

func (*baseEngine) getWebhookUrl(service v1alpha2.AssetWebhookService) string {
	return fmt.Sprintf("http://%s.%s.svc.cluster.local%s", service.Name, service.Namespace, service.Endpoint)
}
