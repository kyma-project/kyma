package webhook

import (
	"encoding/json"
	"fmt"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha1"
	"github.com/pkg/errors"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	"path/filepath"
)

type baseWebhook struct{}

func (*baseWebhook) parseMetadata(metadata *runtime.RawExtension) *json.RawMessage {
	if nil == metadata {
		return nil
	}

	result := json.RawMessage(metadata.Raw)
	return &result
}

func (*baseWebhook) readFiles(basePath string, files []string, reader func(filename string) ([]byte, error)) (map[string]string, error) {
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

func (*baseWebhook) writeFiles(basePath string, content map[string]string) error {
	for key, value := range content {
		path := filepath.Join(basePath, key)
		err := ioutil.WriteFile(path, []byte(value), os.ModePerm)
		if err != nil {
			return errors.Wrapf(err, "while writing file %s", path)
		}
	}

	return nil
}

func (*baseWebhook) getWebhookUrl(service v1alpha1.AssetWebhookService) string {
	return fmt.Sprintf("http://%s.%s.svc.cluster.local%s", service.Name, service.Namespace, service.Endpoint)
}
