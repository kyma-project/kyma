package externalapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/event-service/internal/events/registered"
	"github.com/kyma-project/kyma/components/event-service/internal/events/registered/mocks"
	"github.com/kyma-project/kyma/components/event-service/internal/httperrors"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActiveEventsHandler_GetActiveEvents(t *testing.T) {

	url := "https://gateway.kyma.local/test-app/v1/activeevents"

	t.Run("Should return active events and http code 200", func(t *testing.T) {
		//given
		expectedResponse := registered.ActiveEvents{
			Events: []string{"topic1, topic2"},
		}

		eventsClient := &mocks.EventsClient{}

		eventsClient.On("GetActiveEvents", "test-app").Return(expectedResponse, nil)

		handler := NewActiveEventsHandler(eventsClient)

		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)
		req = mux.SetURLVars(req, map[string]string{"application": "test-app"})
		rr := httptest.NewRecorder()

		//when
		handler.GetActiveEvents(rr, req)

		//then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var response registered.ActiveEvents
		err = json.Unmarshal(responseBody, &response)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, expectedResponse, response)
	})

	t.Run("Should return http code 500 when error occurs", func(t *testing.T) {
		//given
		errorMsg := "Some error"
		eventsClient := &mocks.EventsClient{}

		eventsClient.On("GetActiveEvents", "test-app").Return(registered.ActiveEvents{}, errors.New(errorMsg))

		handler := NewActiveEventsHandler(eventsClient)

		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)
		req = mux.SetURLVars(req, map[string]string{"application": "test-app"})
		rr := httptest.NewRecorder()

		//when
		handler.GetActiveEvents(rr, req)

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
