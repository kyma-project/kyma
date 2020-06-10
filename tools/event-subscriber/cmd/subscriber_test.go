package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockevent struct {
	Id               string `json:"id"`
	Source           string `json:"source"`
	Specversion      string `json:"specversion"`
	Type             string `json:"type"`
	Eventtypeversion string `json:"eventtypeversion,omitempty"`
}

func Test_CheckCounter(t *testing.T) {
	state := SubscriberState{}
	router := state.setupRouter()
	w := httptest.NewRecorder()

	req := httptest.NewRequest("POST", "/", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	req = httptest.NewRequest("GET", "/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	var cr counterResponse
	err := json.Unmarshal(w.Body.Bytes(), &cr)
	assert.NoError(t, err)
	assert.Equal(t, 1, cr.Counter)

	req = httptest.NewRequest("POST", "/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	req = httptest.NewRequest("GET", "/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	err = json.Unmarshal(w.Body.Bytes(), &cr)
	assert.NoError(t, err)
	assert.Equal(t, 2, cr.Counter)

}

func Test_ResetCounter(t *testing.T) {
	state := SubscriberState{}
	router := state.setupRouter()
	w := httptest.NewRecorder()

	req := httptest.NewRequest("POST", "/", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	req = httptest.NewRequest("GET", "/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	var cr counterResponse
	err := json.Unmarshal(w.Body.Bytes(), &cr)
	assert.NoError(t, err)
	assert.Equal(t, 1, cr.Counter)

	req = httptest.NewRequest("DELETE", "/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	req = httptest.NewRequest("POST", "/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	req = httptest.NewRequest("GET", "/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	err = json.Unmarshal(w.Body.Bytes(), &cr)
	assert.NoError(t, err)
	assert.Equal(t, 1, cr.Counter)
}

func Test_SendCE(t *testing.T) {
	type tableTest struct {
		name       string
		event      mockevent
		wantStatus int
	}
	tt := []tableTest{
		{
			name: "Send event with id",
			event: mockevent{
				Id:               "abc",
				Source:           "a",
				Specversion:      "1.0",
				Type:             "a",
				Eventtypeversion: "v1",
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "Send event without eventtypeversion",
			event: mockevent{
				Id:          "abc",
				Source:      "a",
				Specversion: "1.0",
				Type:        "a",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Send event without id",
			event: mockevent{
				Source:           "a",
				Specversion:      "1.0",
				Type:             "a",
				Eventtypeversion: "v1",
			},
			wantStatus: http.StatusBadRequest,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			state := SubscriberState{}
			router := state.setupRouter()
			event, err := json.Marshal(tc.event)
			require.NoError(t, err)
			req := httptest.NewRequest("POST", "/ce", bytes.NewReader(event))
			req.Header.Set("Content-Type", "application/cloudevents+json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			res := w.Result()
			assert.Equal(t, tc.wantStatus, res.StatusCode)

		})
	}
}

func Test_CheckCEUUID(t *testing.T) {
	type tabletest struct {
		name       string
		event      mockevent
		checkPath  string
		wantStatus int
		hasBody    bool
	}
	tt := []tabletest{
		{
			name: "Event received uuid",
			event: mockevent{
				Id:               "abc",
				Source:           "a",
				Specversion:      "1.0",
				Type:             "a",
				Eventtypeversion: "v1",
			},
			checkPath:  "by-uuid/abc",
			wantStatus: http.StatusOK,
			hasBody:    true,
		},
		{
			name: "Event not received uuid",
			event: mockevent{
				Id:               "abc",
				Source:           "a",
				Specversion:      "1.0",
				Type:             "a",
				Eventtypeversion: "v1",
			},
			checkPath:  "by-uuid/foo",
			wantStatus: http.StatusNoContent,
			hasBody:    false,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			state := SubscriberState{}
			router := state.setupRouter()
			event, err := json.Marshal(tc.event)
			require.NoError(t, err)
			req := httptest.NewRequest("POST", "/ce", bytes.NewReader(event))
			req.Header.Set("Content-Type", "application/cloudevents+json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			res := w.Result()
			assert.Equal(t, http.StatusOK, res.StatusCode)

			var ce cloudevents.Event
			w = httptest.NewRecorder()
			req = httptest.NewRequest("GET", "/ce/"+tc.checkPath, nil)
			router.ServeHTTP(w, req)
			res = w.Result()
			assert.Equal(t, tc.wantStatus, res.StatusCode)
			body, err := ioutil.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, tc.hasBody, len(body) > 0)
			if tc.hasBody {
				err = json.Unmarshal(body, &ce)
				assert.NoError(t, err)
			}
		})
	}
}

func Test_CheckCE(t *testing.T) {
	type tabletest struct {
		name       string
		event      mockevent
		checkPath  string
		wantStatus int
	}
	tt := []tabletest{
		{
			name: "Event received",
			event: mockevent{
				Id:               "abc",
				Source:           "a",
				Specversion:      "1.0",
				Type:             "a",
				Eventtypeversion: "v1",
			},
			checkPath:  "a/a/v1",
			wantStatus: http.StatusOK,
		},
		{
			name: "Event not received",
			event: mockevent{
				Id:               "abc",
				Source:           "a",
				Specversion:      "1.0",
				Type:             "a",
				Eventtypeversion: "v1",
			},
			checkPath:  "a/b/v1",
			wantStatus: http.StatusNoContent,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			state := SubscriberState{}
			router := state.setupRouter()
			event, err := json.Marshal(tc.event)
			require.NoError(t, err)
			req := httptest.NewRequest("POST", "/ce", bytes.NewReader(event))
			req.Header.Set("Content-Type", "application/cloudevents+json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			res := w.Result()
			assert.Equal(t, http.StatusOK, res.StatusCode)

			var ce []cloudevents.Event
			w = httptest.NewRecorder()
			req = httptest.NewRequest("GET", "/ce/"+tc.checkPath, nil)
			router.ServeHTTP(w, req)
			res = w.Result()
			assert.Equal(t, tc.wantStatus, res.StatusCode)
			body, err := ioutil.ReadAll(res.Body)
			require.NoError(t, err)
			err = json.Unmarshal(body, &ce)
			assert.NoError(t, err)

		})
	}
}
