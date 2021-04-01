package busola

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"log"

	"github.com/ulikunitz/xz/lzma"
	"github.com/vmihailenco/msgpack/v5"
)

func encodeInitString(data string) string {
	// unmarshall to map[string]interface
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &jsonMap)
	if err != nil {
		panic(err)
	}

	// pack
	packed, _ := msgpack.Marshal(jsonMap)

	// compress
	lzmaString := lzmaEncode(packed)

	// encode
	encoded := base64.RawURLEncoding.EncodeToString([]byte(lzmaString))

	return encoded
}

// TODO: return error
func lzmaEncode(data []byte) string {
	b := &bytes.Buffer{}
	w, err := lzma.NewWriter(b)
	if err != nil {
		log.Fatal(err)
	}
	if _, err = w.Write(data); err != nil {
		log.Fatal(err)
	}
	if err = w.Close(); err != nil {
		log.Fatal(err)
	}
	return b.String()
}
