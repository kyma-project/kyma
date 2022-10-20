package testing

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"

	"golang.org/x/oauth2"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo" // nolint
	. "github.com/onsi/gomega" // nolint

	logf "sigs.k8s.io/controller-runtime/pkg/log"
	// gcp auth etc.
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/client"
	bebtypes "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
)

const (
	TokenURLPath     = "/auth"
	MessagingURLPath = "/messaging"
)

// BEBMock implements a programmable mock for BEB.
type BEBMock struct {
	Requests       *SafeRequests
	Subscriptions  *SafeSubscriptions
	TokenURL       string
	MessagingURL   string
	log            logr.Logger
	AuthResponse   Response
	GetResponse    ResponseWithName
	ListResponse   Response
	CreateResponse Response
	DeleteResponse Response
	server         *httptest.Server
}

func NewBEBMock() *BEBMock {
	logger := logf.Log.WithName("beb mock")
	return &BEBMock{
		Requests:      NewSafeRequests(),
		Subscriptions: NewSafeSubscriptions(),
		log:           logger,
	}
}

type ResponseWithName func(w http.ResponseWriter, subscriptionName string)
type Response func(w http.ResponseWriter)

func (m *BEBMock) Reset() {
	m.log.Info("Initializing requests")
	m.Requests = NewSafeRequests()
	m.Subscriptions = NewSafeSubscriptions()
	m.AuthResponse = BEBAuthResponseSuccess
	m.GetResponse = GetSubscriptionResponse(m)
	m.ListResponse = BEBListSuccess
	m.CreateResponse = BEBCreateSuccess
	m.DeleteResponse = BEBDeleteResponseSuccess
}

func (m *BEBMock) Start() string {
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
			var subscription bebtypes.Subscription
			_ = json.NewDecoder(r.Body).Decode(&subscription)
			m.Requests.PutSubscription(r, subscription)
			key := r.URL.Path + "/" + subscription.Name
			m.Subscriptions.PutSubscription(key, &subscription)
			m.CreateResponse(w)
		case http.MethodGet:
			key := r.URL.Path
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

func (m *BEBMock) Stop() {
	m.server.Close()
}

// GetSubscriptionResponse checks if a subscription exists in the mock.
func GetSubscriptionResponse(m *BEBMock) ResponseWithName {
	return func(w http.ResponseWriter, key string) {
		subscriptionSaved := m.Subscriptions.GetSubscription(key)
		if subscriptionSaved != nil {
			subscriptionSaved.SubscriptionStatus = bebtypes.SubscriptionStatusActive
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(*subscriptionSaved)
			Expect(err).ShouldNot(HaveOccurred())
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

// BEBAuthResponseSuccess writes a oauth2 authentication Response to the writer for the happy-path.
func BEBAuthResponseSuccess(w http.ResponseWriter) {
	token := oauth2.Token{
		AccessToken:  "some-token",
		TokenType:    "",
		RefreshToken: "",
	}
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(token)
	Expect(err).ShouldNot(HaveOccurred())
}

// BEBCreateSuccess writes a Response to the writer for the happy-path of creating a BEB subscription.
func BEBCreateSuccess(w http.ResponseWriter) {
	w.WriteHeader(http.StatusAccepted)
	response := bebtypes.CreateResponse{
		Response: bebtypes.Response{
			StatusCode: http.StatusAccepted,
			Message:    "",
		},
		Href: "",
	}
	err := json.NewEncoder(w).Encode(response)
	Expect(err).ShouldNot(HaveOccurred())
}

// BEBGetSuccess writes a Response to the writer for the happy-path of getting a BEB subscription.
func BEBGetSuccess(w http.ResponseWriter, name string) {
	w.WriteHeader(http.StatusOK)
	s := bebtypes.Subscription{
		Name:               name,
		SubscriptionStatus: bebtypes.SubscriptionStatusActive,
	}
	err := json.NewEncoder(w).Encode(s)
	Expect(err).ShouldNot(HaveOccurred())
}

// BEBListSuccess writes a Response to the writer for the happy-path of listing a BEB subscription.
func BEBListSuccess(w http.ResponseWriter) {
	w.WriteHeader(http.StatusAccepted)
	response := bebtypes.Response{
		StatusCode: http.StatusOK,
		Message:    "",
	}
	err := json.NewEncoder(w).Encode(response)
	Expect(err).ShouldNot(HaveOccurred())
}

// BEBDeleteResponseSuccess writes a Response to the writer for the happy-path of deleting a BEB subscription.
func BEBDeleteResponseSuccess(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// IsBEBSubscriptionCreate determines if the http request is creating a BEB subscription.
func IsBEBSubscriptionCreate(r *http.Request) bool {
	return r.Method == http.MethodPost && strings.Contains(r.RequestURI, client.CreateURL)
}

// IsBEBSubscriptionDelete determines if the http request is deleting a BEB subscription.
func IsBEBSubscriptionDelete(r *http.Request) bool {
	return r.Method == http.MethodDelete && strings.Contains(r.RequestURI, MessagingURLPath)
}

// GetRestAPIObject gets the name of the involved object in a REST url.
// e.g. "/messaging/events/subscriptions/{subscriptionName}" => "{subscriptionName}".
func GetRestAPIObject(u *url.URL) string {
	return path.Base(u.Path)
}
