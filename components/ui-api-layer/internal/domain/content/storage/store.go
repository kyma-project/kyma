package storage

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/golang/glog"

	"github.com/pkg/errors"
	"encoding/json"
)

type notification struct {
	parent    string
	filename  string
	eventType string
}

type store struct {
	bucketName      string
	externalAddress string
	assetFolder     string
	client          client
	assetsRegexp    *regexp.Regexp
}

func newStore(client client, bucketName, externalAddress, assetFolder string) *store {
	pattern := fmt.Sprintf(`"%s/|"\./%s/`, assetFolder, assetFolder)

	return &store{
		bucketName:      bucketName,
		externalAddress: externalAddress,
		assetFolder:     assetFolder,
		client:          client,
		assetsRegexp:    regexp.MustCompile(pattern),
	}
}

func (s *store) Content(id string) (*Content, bool, error) {
	content := new(Content)
	exists, err := s.object(id, "content.json", content)
	if exists {
		s.prepareDocs(content, id)
	}

	return content, exists, err
}

func (s *store) ApiSpec(id string) (*ApiSpec, bool, error) {
	apiSpec := new(ApiSpec)
	exists, err := s.object(id, "apiSpec.json", apiSpec)

	return apiSpec, exists, err
}

func (s *store) OpenApiSpec(id string) (*OpenApiSpec, bool, error) {
	openApiSpec := new(OpenApiSpec)
	exists, err := s.object(id, "apiSpec.json", openApiSpec)

	return openApiSpec, exists, err
}

func (s *store) ODataSpec(id string) (*ODataSpec, bool, error) {
	odataSpec := new(ODataSpec)
	exists, err := s.object(id, "apiSpec.json", odataSpec)

	return odataSpec, exists, err
}

func (s *store) AsyncApiSpec(id string) (*AsyncApiSpec, bool, error) {
	asyncApiSpec := new(AsyncApiSpec)
	exists, err := s.object(id, "asyncApiSpec.json", asyncApiSpec)

	return asyncApiSpec, exists, err
}

func (s *store) NotificationChannel(stop <-chan struct{}) <-chan notification {
	return s.client.NotificationChannel(s.bucketName, stop)
}

func (s *store) object(id, filename string, value interface{}) (bool, error) {
	objectName := fmt.Sprintf("%s/%s", id, filename)
	reader, err := s.client.Object(s.bucketName, objectName)
	defer func() {
		err := reader.Close()
		if err != nil {
			glog.Errorf("Unable to close object reader. Cause: %v", err)
		}
	}()
	if err != nil {
		return false, errors.Wrapf(err, "while getting object `%s`", objectName)
	}

	exists, err := s.decode(reader, value)
	if err != nil || !exists {
		return false, errors.Wrapf(err, "while decoding object `%s`", objectName)
	}

	return true, nil
}

func (s *store) prepareDocs(content *Content, id string) {
	if content == nil {
		return
	}

	docsObj, exists := content.Raw["docs"]
	if !exists {
		return
	}

	docs, ok := docsObj.([]interface{})
	if !ok {
		return
	}

	var result []interface{}
	for _, v := range docs {
		replaced := s.replaceAssetsAddress(v, id)
		if replaced != nil {
			result = append(result, replaced)
		}
	}

	content.Raw["docs"] = result
}

func (s *store) replaceAssetsAddress(in interface{}, id string) interface{} {
	if in == nil {
		return in
	}

	doc, ok := in.(map[string]interface{})
	if !ok {
		return in
	}

	result := make(map[string]interface{})
	for k, v := range doc {
		value, ok := v.(string)
		if ok {
			replaced := strings.Replace(value, "{PLACEHOLDER_APP_RESOURCES_BASE_URI}", s.externalAddress, -1)
			address := fmt.Sprintf(`"%s/%s/%s/%s/`, s.externalAddress, s.bucketName, id, s.assetFolder)
			result[k] = s.assetsRegexp.ReplaceAllString(replaced, address)
		} else {
			result[k] = v
		}
	}

	return result
}

func (s *store) decode(reader io.Reader, value interface{}) (bool, error) {
	err := json.NewDecoder(reader).Decode(value)

	if err != nil {
		if ok := s.client.IsNotExistsError(err); ok {
			return false, nil
		}
		if ok := s.client.IsInvalidBeginningCharacterError(err); ok {
			return s.decodeOData(reader, value)
		}

		return false, err
	}
	return true, nil
}

func (s *store) decodeOData(reader io.Reader, value interface{}) (bool, error) {
	odata, ok := value.(*ODataSpec)
	if !ok {
		return false, nil
	}

	err := odata.UnmarshalXML(reader)
	if err != nil {
		return false, err
	}

	return true, nil
}
