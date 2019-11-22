package passport

import (
	"github.com/kyma-project/kyma/components/application-gateway/internal/proxy/passport/db"
	"net/http"
)

type RequestEnricher struct {
	db *db.DB
}

func New(redisURL string) *RequestEnricher {
	database := db.New(redisURL)
	return &RequestEnricher{db: database}
}
func AnnotatePassportHeaders(request *http.Request) *http.Request {
	return request
}
