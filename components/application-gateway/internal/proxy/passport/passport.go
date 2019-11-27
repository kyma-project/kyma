package passport

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/kyma-project/kyma/components/application-gateway/internal/proxy/passport/db"
)

type RequestEnricher struct {
	db *db.DB
}

func New(redisURL string) *RequestEnricher {
	database := db.New(redisURL)
	return &RequestEnricher{db: database}
}
func (re *RequestEnricher) AnnotatePassportHeaders(request *http.Request, storageKeyName string) *http.Request {
	jsonString, err := re.db.GET(request.Header.Get(storageKeyName))
	if err != nil || jsonString == "" {
		return request
	}

	passportHeaders, err := to(jsonString)
	if err != nil {
		return request
	}

	for k, v := range passportHeaders {
		request.Header.Set(k, v)
	}
	return request
}

func to(jsonString string) (map[string]string, error) {
	passportHeadersMap := make(map[string]string)

	unquotedString := tryUnquote(jsonString)

	err := json.Unmarshal([]byte(unquotedString), &passportHeadersMap)
	if err != nil {
		log.Printf("error when unmarshalling jsonString {%s} to map %+v", unquotedString, err)
		return nil, err
	}

	log.Printf("unmarshalled passport headers %+v", passportHeadersMap)

	return passportHeadersMap, nil
}

func tryUnquote(jsonString string) string {
	unquotedString, err := strconv.Unquote(jsonString)
	if err != nil {
		log.Printf("error when unquoting jsonString %s to map %+v", jsonString, err)
		return jsonString
	}
	return unquotedString
}
