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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

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

// BebMock implements a programmable mock for BEB
type BebMock struct {
	Requests       map[*http.Request]interface{}
	Subscriptions  map[string]*bebtypes.Subscription
	TokenURL       string
	MessagingURL   string
	BebConfig      *config.Config
	log            logr.Logger
	AuthResponse   response
	GetResponse    responseWithName
	ListResponse   response
	CreateResponse response
	DeleteResponse response
}

func NewBebMock(bebConfig *config.Config) *BebMock {
	logger := logf.Log.WithName("beb mock")
	return &BebMock{
		nil, nil, "", "", bebConfig,
		logger,
		nil, nil, nil, nil, nil,
	}
}

type responseWithName func(w http.ResponseWriter, subscriptionName string)
type response func(w http.ResponseWriter)

func (m *BebMock) Reset() {
	m.log.Info("Initializing requests")
	m.Requests = make(map[*http.Request]interface{})
	m.Subscriptions = make(map[string]*bebtypes.Subscription)
	m.AuthResponse = nil
	m.GetResponse = nil
	m.ListResponse = nil
	m.CreateResponse = nil
	m.DeleteResponse = nil
}

func (m *BebMock) Start() string {
	m.Reset()

	// implementation based on https://pages.github.tools.sap/KernelServices/APIDefinitions/?urls.primaryName=Business%20Event%20Bus%20-%20CloudEvents
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer GinkgoRecover()

		// store request
		m.Requests[r] = nil

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
		if r.Method == http.MethodPost && r.RequestURI == TokenURLPath {
			if m.AuthResponse != nil {
				m.AuthResponse(w)
			} else {
				BebAuthResponseSuccess(w)
			}
			return
		}
		// messaging API request
		if strings.HasPrefix(r.RequestURI, MessagingURLPath) {
			switch r.Method {
			case http.MethodDelete:
				key := r.URL.Path
				delete(m.Subscriptions, key)
				if m.DeleteResponse != nil {
					m.DeleteResponse(w)
				} else {
					BebDeleteResponseSuccess(w)
				}
			case http.MethodPost:
				var subscription bebtypes.Subscription
				_ = json.NewDecoder(r.Body).Decode(&subscription)
				m.Requests[r] = subscription
				key := r.URL.Path + "/" + subscription.Name
				m.Subscriptions[key] = &subscription
				if m.CreateResponse != nil {
					m.CreateResponse(w)
				} else {
					BebCreateSuccess(w)
				}
			case http.MethodGet:
				switch r.RequestURI {
				case m.BebConfig.ListURL:
					if m.ListResponse != nil {
						m.ListResponse(w)
					} else {
						BebListSuccess(w)
					}
				// get on a single subscription
				default:
					key := r.URL.Path
					subscription := m.Subscriptions[key]
					if subscription != nil {
						subscription.SubscriptionStatus = bebtypes.SubscriptionStatusActive
						w.WriteHeader(http.StatusOK)
						err := json.NewEncoder(w).Encode(*subscription)
						Expect(err).ShouldNot(HaveOccurred())
					} else {
						w.WriteHeader(http.StatusNotFound)
					}
					// TODO make it work for all other BEB mock tests (see reconciler_test)
					/*  old stuff
					var subscription bebtypes.Subscription
					_ = json.NewDecoder(r.Body).Decode(&subscription)
					m.Requests[r] = subscription

					parsedUrl, err := url.Parse(r.RequestURI)
					Expect(err).ShouldNot(HaveOccurred())
					subscriptionName := parsedUrl.Path
					if m.GetResponse != nil {
						m.GetResponse(w, subscriptionName)
					} else {
						BebGetSuccess(w, subscriptionName)
					}
					*/
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

// BebAuthResponseSuccess writes a oauth2 authentication response to the writer for the happy-path.
func BebAuthResponseSuccess(w http.ResponseWriter) {
	token := oauth2.Token{
		AccessToken:  "some-token",
		TokenType:    "",
		RefreshToken: "",
	}
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(token)
	Expect(err).ShouldNot(HaveOccurred())
}

// BebCreateSuccess writes a response to the writer for the happy-path of creating a BEB subscription.
func BebCreateSuccess(w http.ResponseWriter) {
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

// BebGetSuccess writes a response to the writer for the happy-path of getting a BEB subscription.
func BebGetSuccess(w http.ResponseWriter, name string) {
	w.WriteHeader(http.StatusOK)
	s := bebtypes.Subscription{
		Name:               name,
		SubscriptionStatus: bebtypes.SubscriptionStatusActive,
	}
	err := json.NewEncoder(w).Encode(s)
	Expect(err).ShouldNot(HaveOccurred())
}

// BebListSuccess writes a response to the writer for the happy-path of listing a BEB subscription.
func BebListSuccess(w http.ResponseWriter) {
	w.WriteHeader(http.StatusAccepted)
	response := bebtypes.Response{
		StatusCode: http.StatusOK,
		Message:    "",
	}
	err := json.NewEncoder(w).Encode(response)
	Expect(err).ShouldNot(HaveOccurred())
}

// BebDeleteResponseSuccess writes a response to the writer for the happy-path of deleting a BEB subscription.
func BebDeleteResponseSuccess(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// IsBebSubscriptionCreate determines if the http request is creating a BEB subscription.
func IsBebSubscriptionCreate(r *http.Request, bebConfig config.Config) bool {
	return r.Method == http.MethodPost && strings.Contains(bebConfig.CreateURL, r.RequestURI)
}

// IsBebSubscriptionCreate determines if the http request is deleting a BEB subscription.
func IsBebSubscriptionDelete(r *http.Request) bool {
	return r.Method == http.MethodDelete && strings.Contains(r.RequestURI, MessagingURLPath)
}

// GetRestAPIObject gets the name of the involved object in a REST url
// e.g. "/messaging/events/subscriptions/{subscriptionName}" => "{subscriptionName}"
func GetRestAPIObject(u *url.URL) string {
	return path.Base(u.Path)
}
