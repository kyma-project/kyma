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

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/config"
	bebtypes "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
)

const (
	TokenURLPath     = "/auth"
	MessagingURLPath = "/messaging"
)

// BEBMock implements a programmable mock for BEB
type BEBMock struct {
	Requests       *SafeRequests
	Subscriptions  *SafeSubscriptions
	TokenURL       string
	MessagingURL   string
	BEBConfig      *config.Config
	log            logr.Logger
	AuthResponse   response
	GetResponse    responseWithName
	ListResponse   response
	CreateResponse response
	DeleteResponse response
}

func NewBEBMock(bebConfig *config.Config) *BEBMock {
	logger := logf.Log.WithName("beb mock")
	return &BEBMock{
		NewSafeRequests(), NewSafeSubscriptions(), "", "", bebConfig,
		logger,
		nil, nil, nil, nil, nil,
	}
}

type responseWithName func(w http.ResponseWriter, subscriptionName string)
type response func(w http.ResponseWriter)

func (m *BEBMock) Reset() {
	m.log.Info("Initializing requests")
	m.Requests = NewSafeRequests()
	m.Subscriptions = NewSafeSubscriptions()
	m.AuthResponse = nil
	m.GetResponse = nil
	m.ListResponse = nil
	m.CreateResponse = nil
	m.DeleteResponse = nil
}

func (m *BEBMock) Start() string {
	m.Reset()

	// implementation based on https://pages.github.tools.sap/KernelServices/APIDefinitions/?urls.primaryName=Business%20Event%20Bus%20-%20CloudEvents
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

		// oauth2 request
		if r.Method == http.MethodPost && strings.HasPrefix(r.RequestURI, TokenURLPath) {
			if m.AuthResponse != nil {
				m.AuthResponse(w)
			} else {
				BEBAuthResponseSuccess(w)
			}
			return
		}
		// messaging API request
		if strings.HasPrefix(r.RequestURI, MessagingURLPath) {
			switch r.Method {
			case http.MethodDelete:
				key := r.URL.Path
				m.Subscriptions.DeleteSubscription(key)
				if m.DeleteResponse != nil {
					m.DeleteResponse(w)
				} else {
					BEBDeleteResponseSuccess(w)
				}
			case http.MethodPost:
				var subscription bebtypes.Subscription
				_ = json.NewDecoder(r.Body).Decode(&subscription)
				m.Requests.PutSubscription(r, subscription)
				key := r.URL.Path + "/" + subscription.Name
				m.Subscriptions.PutSubscription(key, &subscription)
				if m.CreateResponse != nil {
					m.CreateResponse(w)
				} else {
					BEBCreateSuccess(w)
				}
			case http.MethodGet:
				switch r.RequestURI {
				case m.BEBConfig.ListURL:
					if m.ListResponse != nil {
						m.ListResponse(w)
					} else {
						BEBListSuccess(w)
					}
				// get on a single subscription
				default:
					key := r.URL.Path
					if m.GetResponse != nil {
						m.GetResponse(w, key)
					} else {
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
				return
			default:
				w.WriteHeader(http.StatusNotImplemented)
			}
			return
		}
	}))
	uri := ts.URL

	return uri
}

// BEBAuthResponseSuccess writes a oauth2 authentication response to the writer for the happy-path.
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

// BEBCreateSuccess writes a response to the writer for the happy-path of creating a BEB subscription.
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

// BEBGetSuccess writes a response to the writer for the happy-path of getting a BEB subscription.
func BEBGetSuccess(w http.ResponseWriter, name string) {
	w.WriteHeader(http.StatusOK)
	s := bebtypes.Subscription{
		Name:               name,
		SubscriptionStatus: bebtypes.SubscriptionStatusActive,
	}
	err := json.NewEncoder(w).Encode(s)
	Expect(err).ShouldNot(HaveOccurred())
}

// BEBListSuccess writes a response to the writer for the happy-path of listing a BEB subscription.
func BEBListSuccess(w http.ResponseWriter) {
	w.WriteHeader(http.StatusAccepted)
	response := bebtypes.Response{
		StatusCode: http.StatusOK,
		Message:    "",
	}
	err := json.NewEncoder(w).Encode(response)
	Expect(err).ShouldNot(HaveOccurred())
}

// BEBDeleteResponseSuccess writes a response to the writer for the happy-path of deleting a BEB subscription.
func BEBDeleteResponseSuccess(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// IsBEBSubscriptionCreate determines if the http request is creating a BEB subscription.
func IsBEBSubscriptionCreate(r *http.Request, bebConfig config.Config) bool {
	return r.Method == http.MethodPost && strings.Contains(bebConfig.CreateURL, r.RequestURI)
}

// IsBEBSubscriptionDelete determines if the http request is deleting a BEB subscription.
func IsBEBSubscriptionDelete(r *http.Request) bool {
	return r.Method == http.MethodDelete && strings.Contains(r.RequestURI, MessagingURLPath)
}

// GetRestAPIObject gets the name of the involved object in a REST url
// e.g. "/messaging/events/subscriptions/{subscriptionName}" => "{subscriptionName}"
func GetRestAPIObject(u *url.URL) string {
	return path.Base(u.Path)
}
