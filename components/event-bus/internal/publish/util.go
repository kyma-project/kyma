package publish

import (
	"encoding/json"
	"net/http"

	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
)

//SendJSONError sends an HTTP response containing a JSON error
func SendJSONError(w http.ResponseWriter, err *api.Error) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader((*err).Status)
	return json.NewEncoder(w).Encode(*err)
}
