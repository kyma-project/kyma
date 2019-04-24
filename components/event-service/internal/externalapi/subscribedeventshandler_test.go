package externalapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/event-service/internal/events/subscribed"
	"github.com/kyma-project/kyma/components/event-service/internal/events/subscribed/mocks"
	"github.com/kyma-project/kyma/components/event-service/internal/httperrors"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActiveEventsHandler_GetActiveEvents(t *testing.T) {

	url := "https://gateway.kyma.local/test-app/v1/activeevents"

	t.Run("Should return active events and http code 200", func(t *testing.T) {
		//given
		expectedResponse := subscribed.Events{
			Events: []subscribed.Event{{Name: "topic1", Version: "v1"}, {Name: "topic2", Version: "v1"}},
		}

		eventsClient := &mocks.EventsClient{}

		eventsClient.On("GetSubscribedEvents", "test-app").Return(expectedResponse, nil)

		handler := NewActiveEventsHandler(eventsClient)

		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)
		req = mux.SetURLVars(req, map[string]string{"application": "test-app"})
		rr := httptest.NewRecorder()

		//when
		handler.GetSubscribedEvents(rr, req)

		//then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var response subscribed.Events
		err = json.Unmarshal(responseBody, &response)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, expectedResponse, response)
	})

	t.Run("Should return http code 500 when error occurs", func(t *testing.T) {
		//given
		errorMsg := "Some error"
		eventsClient := &mocks.EventsClient{}

		eventsClient.On("GetSubscribedEvents", "test-app").Return(subscribed.Events{}, errors.New(errorMsg))

		handler := NewActiveEventsHandler(eventsClient)

		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)
		req = mux.SetURLVars(req, map[string]string{"application": "test-app"})
		rr := httptest.NewRecorder()

		//when
		handler.GetSubscribedEvents(rr, req)

		//then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var response httperrors.ErrorResponse
		err = json.Unmarshal(responseBody, &response)
		require.NoError(t, err)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Equal(t, errorMsg, response.Error)
	})
}
