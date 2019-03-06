package proxy

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/apiserver-proxy/internal/authn"
	"github.com/kyma-project/kyma/components/apiserver-proxy/internal/authz"

	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
)

const KUBERNETES_SERVICE = "kubernetes.default"

// Config holds proxy authorization and authentication settings
type Config struct {
	Authentication *authn.AuthnConfig
	Authorization  *authz.Config
}

type kubeRBACProxy struct {

	// authenticator identifies the user for requests to kube-rbac-proxy
	authenticator.Request
	// authorizerAttributeGetter builds authorization.Attributes for a request to kube-rbac-proxy
	authorizer.RequestAttributesGetter
	// authorizer determines whether a given authorization.Attributes is allowed
	authorizer.Authorizer
	// config for kube-rbac-proxy
	Config Config
}

// New creates an authenticator, an authorizer, and a matching authorizer attributes getter compatible with the kube-rbac-proxy
func New(config Config, authorizer authorizer.Authorizer, authenticator authenticator.Request) *kubeRBACProxy {
	return &kubeRBACProxy{authenticator, newKubeRBACProxyAuthorizerAttributesGetter(config.Authorization), authorizer, config}
}

// Handle authenticates the client and authorizes the request.
// If the authn fails, a 401 error is returned. If the authz fails, a 403 error is returned
func (h *kubeRBACProxy) Handle(w http.ResponseWriter, req *http.Request) bool {
	// Authenticate
	u, ok, err := h.AuthenticateRequest(req)
	if err != nil {
		glog.Errorf("Unable to authenticate the request due to an error: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return false
	}
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return false
	}

	// Get authorization attributes
	attrs := h.GetRequestAttributes(u, req)

	// Authorize
	authorized, _, err := h.Authorize(attrs)
	if err != nil {
		msg := fmt.Sprintf("Authorization error (user=%s, verb=%s, resource=%s, subresource=%s)", u.GetName(), attrs.GetVerb(), attrs.GetResource(), attrs.GetSubresource())
		glog.Errorf(msg, err)
		http.Error(w, msg, http.StatusInternalServerError)
		return false
	}
	if authorized != authorizer.DecisionAllow {
		msg := fmt.Sprintf("Forbidden (user=%s, verb=%s, resource=%s, subresource=%s)", u.GetName(), attrs.GetVerb(), attrs.GetResource(), attrs.GetSubresource())
		glog.V(2).Info(msg)
		http.Error(w, msg, http.StatusForbidden)
		return false
	}

	if h.Config.Authentication.Header.Enabled {
		// Seemingly well-known headers to tell the upstream about user's identity
		// so that the upstream can achieve the original goal of delegating RBAC authn/authz to kube-rbac-proxy
		headerCfg := h.Config.Authentication.Header
		req.Header.Set(headerCfg.UserFieldName, u.GetName())
		req.Header.Set(headerCfg.GroupsFieldName, strings.Join(u.GetGroups(), headerCfg.GroupSeparator))
	}

	return true
}

func newKubeRBACProxyAuthorizerAttributesGetter(authzConfig *authz.Config) authorizer.RequestAttributesGetter {
	return krpAuthorizerAttributesGetter{authzConfig, newRequestInfoResolver()}
}

type krpAuthorizerAttributesGetter struct {
	authzConfig     *authz.Config
	reqInfoResolver *apirequest.RequestInfoFactory
}

// GetRequestAttributes populates authorizer attributes for the requests to kube-rbac-proxy.
func (n krpAuthorizerAttributesGetter) GetRequestAttributes(u user.Info, r *http.Request) authorizer.Attributes {
	apiVerb := ""
	switch r.Method {
	case "POST":
		apiVerb = "create"
	case "GET":
		apiVerb = "get"
	case "PUT":
		apiVerb = "update"
	case "PATCH":
		apiVerb = "patch"
	case "DELETE":
		apiVerb = "delete"
	}

	raf := n.authzConfig.ResourceAttributesFile
	if raf != "" {
		b, err := ioutil.ReadFile(raf)
		if err != nil {
			glog.Fatalf("Failed to read resource-attribute file: %v", err)
		}

		err = yaml.Unmarshal(b, &n.authzConfig.ResourceAttributes)
		if err != nil {
			glog.Fatalf("Failed to parse resource-attribute file content: %v", err)
		}
	}

	requestPath := r.URL.Path
	// Default attributes mirror the API attributes that would allow this access to kube-rbac-proxy
	attrs := authorizer.AttributesRecord{
		User:            u,
		Verb:            apiVerb,
		Namespace:       "",
		APIGroup:        "",
		APIVersion:      "",
		Resource:        "",
		Subresource:     "",
		Name:            "",
		ResourceRequest: false,
		Path:            requestPath,
	}

	//attributes based on configuration loaded from file
	if n.authzConfig.ResourceAttributes != nil {
		attrs = authorizer.AttributesRecord{
			User:            u,
			Verb:            apiVerb,
			Namespace:       n.authzConfig.ResourceAttributes.Namespace,
			APIGroup:        n.authzConfig.ResourceAttributes.APIGroup,
			APIVersion:      n.authzConfig.ResourceAttributes.APIVersion,
			Resource:        n.authzConfig.ResourceAttributes.Resource,
			Subresource:     n.authzConfig.ResourceAttributes.Subresource,
			Name:            n.authzConfig.ResourceAttributes.Name,
			ResourceRequest: true,
		}
	} else {
		// attributes based on request
		reqInfo, err := n.reqInfoResolver.NewRequestInfo(r)

		if err != nil {
			glog.Fatalf("Unable to create request info object. %v", err)
		}

		attrs.User = u
		attrs.Verb = reqInfo.Verb
		attrs.APIGroup = reqInfo.APIGroup
		attrs.APIVersion = reqInfo.APIVersion
		attrs.Name = reqInfo.Name
		attrs.Namespace = reqInfo.Namespace
		attrs.ResourceRequest = reqInfo.IsResourceRequest
		attrs.Resource = reqInfo.Resource
		attrs.Subresource = reqInfo.Subresource
		attrs.Path = reqInfo.Path
	}

	glog.V(5).Infof("kube-rbac-proxy request attributes: attrs=%#v", attrs)

	return attrs
}

// DeepCopy of Proxy Configuration
func (c *Config) DeepCopy() *Config {
	res := &Config{
		Authentication: &authn.AuthnConfig{},
	}

	if c.Authentication != nil {
		res.Authentication = &authn.AuthnConfig{}

		if c.Authentication.X509 != nil {
			res.Authentication.X509 = &authn.X509Config{
				ClientCAFile: c.Authentication.X509.ClientCAFile,
			}
		}

		if c.Authentication.Header != nil {
			res.Authentication.Header = &authn.AuthnHeaderConfig{
				Enabled:         c.Authentication.Header.Enabled,
				UserFieldName:   c.Authentication.Header.UserFieldName,
				GroupsFieldName: c.Authentication.Header.GroupsFieldName,
				GroupSeparator:  c.Authentication.Header.GroupSeparator,
			}
		}
	}

	if c.Authorization != nil {
		if c.Authorization.ResourceAttributes != nil {
			res.Authorization = &authz.Config{
				ResourceAttributes: &authz.ResourceAttributes{
					Namespace:   c.Authorization.ResourceAttributes.Namespace,
					APIGroup:    c.Authorization.ResourceAttributes.APIGroup,
					APIVersion:  c.Authorization.ResourceAttributes.APIVersion,
					Resource:    c.Authorization.ResourceAttributes.Resource,
					Subresource: c.Authorization.ResourceAttributes.Subresource,
					Name:        c.Authorization.ResourceAttributes.Name,
				},
			}
		}
	}

	return res
}

func newRequestInfoResolver() *apirequest.RequestInfoFactory {
	apiPrefixes := sets.NewString("apis", "api")
	legacyAPIPrefixes := sets.NewString("api")

	return &apirequest.RequestInfoFactory{
		APIPrefixes:          apiPrefixes,
		GrouplessAPIPrefixes: legacyAPIPrefixes,
	}
}
