package busola

import (
	"bytes"
	"encoding/base64"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/ulikunitz/xz/lzma"
	"github.com/vmihailenco/msgpack/v5"
)

func encodeInitString(data string) (string, error) {
	// unmarshall to map[string]interface
	jsonMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(data), &jsonMap); err != nil {
		return "", errors.Wrap(err, "while unmarshalling data to map[string]interface")
	}

	// pack
	packed, err := msgpack.Marshal(jsonMap)
	if err != nil {
		return "", errors.Wrap(err, "while marshalling data to msgpack")
	}

	// compress
	lzmaString, err := lzmaEncode(packed)
	if err != nil {
		return "", errors.Wrap(err, "while compressing data using LZMA algorithm")
	}

	// encode
	encoded := base64.RawURLEncoding.EncodeToString([]byte(lzmaString))

	return encoded, nil
}

func lzmaEncode(data []byte) (string, error) {
	b := new(bytes.Buffer)
	w, err := lzma.NewWriter(b)
	if err != nil {
		return "", err
	}
	if _, err = w.Write(data); err != nil {
		return "", err
	}
	if err = w.Close(); err != nil {
		return "", err
	}
	return b.String(), nil
}
