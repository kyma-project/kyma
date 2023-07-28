package testing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo" //nolint:revive,stylecheck // using . import for convenience
	. "github.com/onsi/gomega" //nolint:revive,stylecheck // using . import for convenience
	"golang.org/x/oauth2"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	// gcp auth etc.
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/client"
	eventmeshtypes "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
)

const (
	TokenURLPath     = "/auth"
	MessagingURLPath = "/messaging"
)

// EventMeshMock implements a programmable mock for EventMesh.
type EventMeshMock struct {
	Requests            *SafeRequests
	Subscriptions       *SafeSubscriptions
	TokenURL            string
	MessagingURL        string
	log                 logr.Logger
	AuthResponse        Response
	GetResponse         ResponseWithName
	ListResponse        Response
	CreateResponse      Response
	UpdateResponse      ResponseUpdateReq
	UpdateStateResponse ResponseUpdateStateReq
	DeleteResponse      Response
	server              *httptest.Server
	ResponseOverrides   *EventMeshMockResponseOverride
}

type EventMeshMockResponseOverride struct {
	CreateResponse map[string]ResponseWithSub
	GetResponse    map[string]ResponseWithName
}

func NewEventMeshMock() *EventMeshMock {
	logger := logf.Log.WithName("beb mock")
	return &EventMeshMock{
		Requests:          NewSafeRequests(),
		Subscriptions:     NewSafeSubscriptions(),
		log:               logger,
		ResponseOverrides: NewEventMeshMockResponseOverride(),
	}
}

func NewEventMeshMockResponseOverride() *EventMeshMockResponseOverride {
	return &EventMeshMockResponseOverride{
		CreateResponse: map[string]ResponseWithSub{},
		GetResponse:    map[string]ResponseWithName{},
	}
}

type ResponseUpdateReq func(w http.ResponseWriter, key string, webhookAuth *eventmeshtypes.WebhookAuth)
type ResponseUpdateStateReq func(w http.ResponseWriter, key string, state eventmeshtypes.State)
type ResponseWithSub func(w http.ResponseWriter, subscription eventmeshtypes.Subscription)
type ResponseWithName func(w http.ResponseWriter, subscriptionName string)
type Response func(w http.ResponseWriter)

func (m *EventMeshMock) Reset() {
	m.log.Info("Initializing requests")
	m.Requests = NewSafeRequests()
	m.Subscriptions = NewSafeSubscriptions()
	m.AuthResponse = EventMeshAuthResponseSuccess
	m.GetResponse = GetSubscriptionResponse(m)
	m.ListResponse = EventMeshListSuccess
	m.CreateResponse = EventMeshCreateSuccess
	m.DeleteResponse = EventMeshDeleteResponseSuccess
	m.ResponseOverrides = NewEventMeshMockResponseOverride()
	m.UpdateResponse = UpdateSubscriptionResponse(m)
	m.UpdateStateResponse = UpdateSubscriptionStateResponse(m)
}

func (m *EventMeshMock) ResetResponseOverrides() {
	m.log.Info("Resetting response overrides")
	m.ResponseOverrides = NewEventMeshMockResponseOverride()
}

func (m *EventMeshMock) AddCreateResponseOverride(key string, responseFunc ResponseWithSub) {
	m.ResponseOverrides.CreateResponse[key] = responseFunc
}

func (m *EventMeshMock) AddGetResponseOverride(key string, responseFunc ResponseWithName) {
	m.ResponseOverrides.GetResponse[key] = responseFunc
}

func (m *EventMeshMock) Start() string {
	m.Reset()

	// implementation based on https://pages.github.tools.sap/KernelServices/APIDefinitions/?urls.primaryName=Business%20Event%20Bus%20-%20CloudEvents
	mux := http.NewServeMux()

	// oauth2 request
	mux.HandleFunc(TokenURLPath, func(w http.ResponseWriter, r *http.Request) {
		// TODO(k15r): method not allowed/implementd handling
		if r.Method == http.MethodPost {
			m.AuthResponse(w)
		}
	})

	mux.HandleFunc(client.ListURL, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			m.ListResponse(w)
		}
	})

	mux.HandleFunc(MessagingURLPath+"/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			key := r.URL.Path
			m.Subscriptions.DeleteSubscription(key)
			m.DeleteResponse(w)
		case http.MethodPost:
			var subscription eventmeshtypes.Subscription
			_ = json.NewDecoder(r.Body).Decode(&subscription)
			key := r.URL.Path + "/" + subscription.Name
			// check if any response override defined for this subscription
			if overrideFunc, ok := m.ResponseOverrides.CreateResponse[key]; ok {
				overrideFunc(w, subscription)
				return
			}

			// otherwise, use default flow
			m.Requests.PutSubscription(r, subscription)
			m.Subscriptions.PutSubscription(key, &subscription)
			m.CreateResponse(w)
		case http.MethodPatch: // mock update WebhookAuth config
			var subscription eventmeshtypes.Subscription
			err := json.NewDecoder(r.Body).Decode(&subscription)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			key := r.URL.Path // i.e. Path will be `/messaging/events/subscriptions/<name>`
			// save request.
			m.Requests.PutSubscription(r, subscription)
			m.UpdateResponse(w, key, subscription.WebhookAuth)
		case http.MethodPut: // mock pause/resume EventMesh subscription
			var state eventmeshtypes.State
			if err := json.NewDecoder(r.Body).Decode(&state); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// extract get request key from /messaging/events/subscriptions/%s/state
			key := strings.TrimSuffix(r.URL.Path, "/state")
			m.UpdateStateResponse(w, key, state)
		case http.MethodGet:
			key := r.URL.Path
			// check if any response override defined for this subscription
			if overrideFunc, ok := m.ResponseOverrides.GetResponse[key]; ok {
				overrideFunc(w, key)
				return
			}

			// otherwise, use default flow
			m.GetResponse(w, key)
		default:
			w.WriteHeader(http.StatusNotImplemented)
		}
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		m.log.V(1).Info(r.RequestURI)
	})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer GinkgoRecover()

		// store request
		m.Requests.StoreRequest(r)

		description := ""
		reqBytes, err := httputil.DumpRequest(r, true)
		if err == nil {
			description = string(reqBytes)
		}
		m.log.V(1).Info("received request",
			"uri", r.RequestURI,
			"method", r.Method,
			"description", description,
		)

		w.Header().Set("Content-Type", "application/json")
		mux.ServeHTTP(w, r)
	}))
	uri := ts.URL
	m.server = ts
	m.MessagingURL = m.server.URL + MessagingURLPath
	m.TokenURL = m.server.URL + TokenURLPath
	return uri
}

func (m *EventMeshMock) Stop() {
	m.server.Close()
}

// GetSubscriptionResponse checks if a subscription exists in the mock.
func GetSubscriptionResponse(m *EventMeshMock) ResponseWithName {
	return func(w http.ResponseWriter, key string) {
		subscriptionSaved := m.Subscriptions.GetSubscription(key)
		if subscriptionSaved != nil {
			if subscriptionSaved.SubscriptionStatus == "" {
				subscriptionSaved.SubscriptionStatus = eventmeshtypes.SubscriptionStatusActive
			}
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(*subscriptionSaved)
			Expect(err).ShouldNot(HaveOccurred())
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

// UpdateSubscriptionResponse updates the webhook auth of subscription in the mock.
func UpdateSubscriptionResponse(m *EventMeshMock) ResponseUpdateReq {
	return func(w http.ResponseWriter, key string, webhookAuth *eventmeshtypes.WebhookAuth) {
		subscriptionSaved := m.Subscriptions.GetSubscription(key)
		if subscriptionSaved != nil {
			subscriptionSaved.WebhookAuth = webhookAuth
			m.Subscriptions.PutSubscription(key, subscriptionSaved)
			w.WriteHeader(http.StatusNoContent)
			err := json.NewEncoder(w).Encode(*subscriptionSaved)
			Expect(err).ShouldNot(HaveOccurred())
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

// UpdateSubscriptionStateResponse updates the EventMesh subscription status in the mock.
func UpdateSubscriptionStateResponse(m *EventMeshMock) ResponseUpdateStateReq {
	return func(w http.ResponseWriter, key string, state eventmeshtypes.State) {
		if subscription := m.Subscriptions.GetSubscription(key); subscription != nil {
			switch state.Action {
			case eventmeshtypes.StateActionPause:
				{
					subscription.SubscriptionStatus = eventmeshtypes.SubscriptionStatusPaused
				}
			case eventmeshtypes.StateActionResume:
				{
					subscription.SubscriptionStatus = eventmeshtypes.SubscriptionStatusActive
				}
			default:
				{
					panic(fmt.Sprintf("EventMesh subscription status is not supported: %#v", state))
				}
			}

			m.Subscriptions.PutSubscription(key, subscription)
			w.WriteHeader(http.StatusAccepted)

			err := json.NewEncoder(w).Encode(*subscription)
			Expect(err).ShouldNot(HaveOccurred())
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}
}

// EventMeshAuthResponseSuccess writes an oauth2 authentication Response to the writer for the happy-path.
func EventMeshAuthResponseSuccess(w http.ResponseWriter) {
	token := oauth2.Token{
		AccessToken:  "some-token",
		TokenType:    "",
		RefreshToken: "",
	}
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(token)
	Expect(err).ShouldNot(HaveOccurred())
}

// EventMeshCreateSuccess writes a Response to the writer for the happy-path of creating an EventMesh subscription.
func EventMeshCreateSuccess(w http.ResponseWriter) {
	w.WriteHeader(http.StatusAccepted)
	response := eventmeshtypes.CreateResponse{
		Response: eventmeshtypes.Response{
			StatusCode: http.StatusAccepted,
			Message:    "",
		},
		Href: "",
	}
	err := json.NewEncoder(w).Encode(response)
	Expect(err).ShouldNot(HaveOccurred())
}

// EventMeshGetSuccess writes a Response to the writer for the happy-path of getting an EventMesh subscription.
func EventMeshGetSuccess(w http.ResponseWriter, name string) {
	w.WriteHeader(http.StatusOK)
	s := eventmeshtypes.Subscription{
		Name:               name,
		SubscriptionStatus: eventmeshtypes.SubscriptionStatusActive,
	}
	err := json.NewEncoder(w).Encode(s)
	Expect(err).ShouldNot(HaveOccurred())
}

// EventMeshListSuccess writes a Response to the writer for the happy-path of listing a EventMesh subscription.
func EventMeshListSuccess(w http.ResponseWriter) {
	w.WriteHeader(http.StatusAccepted)
	response := eventmeshtypes.Response{
		StatusCode: http.StatusOK,
		Message:    "",
	}
	err := json.NewEncoder(w).Encode(response)
	Expect(err).ShouldNot(HaveOccurred())
}

// EventMeshDeleteResponseSuccess writes a Response to the writer for the happy-path of deleting an EventMesh
// subscription.
func EventMeshDeleteResponseSuccess(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// IsEventMeshSubscriptionCreate determines if the http request is creating an EventMesh subscription.
func IsEventMeshSubscriptionCreate(r *http.Request) bool {
	return r.Method == http.MethodPost && strings.Contains(r.RequestURI, client.CreateURL)
}

// IsEventMeshSubscriptionDelete determines if the http request is deleting a EventMesh subscription.
func IsEventMeshSubscriptionDelete(r *http.Request) bool {
	return r.Method == http.MethodDelete && strings.Contains(r.RequestURI, MessagingURLPath)
}

// GetRestAPIObject gets the name of the involved object in a REST url.
// e.g. "/messaging/events/subscriptions/{subscriptionName}" => "{subscriptionName}".
func GetRestAPIObject(u *url.URL) string {
	return path.Base(u.Path)
}

// CountRequests counts the mock API requests using the given HTTP method and URI.
func (m *EventMeshMock) CountRequests(method, uri string) int {
	count := 0
	m.Requests.ReadEach(func(request *http.Request, payload interface{}) {
		if request.Method != method {
			return
		}
		if request.RequestURI != uri {
			return
		}
		count++
	})
	return count
}
