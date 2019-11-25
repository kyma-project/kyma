package passport

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/kyma-project/kyma/components/application-gateway/internal/proxy/passport/db"
)

type RequestEnricher struct {
	db *db.DB
}

func New(redisURL string) *RequestEnricher {
	database := db.New(redisURL)
	return &RequestEnricher{db: database}
}
func (re *RequestEnricher) AnnotatePassportHeaders(request *http.Request) *http.Request {
	jsonString, err := re.db.GET(request.Header.Get("x-b3-traceid"))
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
	err := json.Unmarshal([]byte(jsonString), &passportHeadersMap)
	if err != nil {
		log.Printf("error when unmarshalling jsonString %s to map", jsonString)
		return nil, err
	}

	log.Printf("unmarshalled passport headers %+v", passportHeadersMap)

	return passportHeadersMap, nil
}
