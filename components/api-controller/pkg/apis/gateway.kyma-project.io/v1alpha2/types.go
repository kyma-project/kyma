package v1alpha2

import (
	"encoding/json"
	"errors"
	"fmt"

	kymaMeta "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/meta/v1"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Api struct {
	k8sMeta.TypeMeta   `json:",inline"`
	k8sMeta.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApiSpec   `json:"spec"`
	Status ApiStatus `json:"status"`
}

type ApiSpec struct {
	Service                    Service `json:"service"`
	Hostname                   string  `json:"hostname"`
	DisableIstioAuthPolicyMTLS *bool   `json:"disableIstioAuthPolicyMTLS,omitempty"`
	AuthenticationEnabled      *bool   `json:"authenticationEnabled,omitempty"`
	// +optional
	Authentication []AuthenticationRule `json:"authentication"`
}

type Service struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

type AuthenticationRule struct {
	Type AuthenticationType `json:"type"`
	Jwt  JwtAuthentication  `json:"jwt"`
}

type AuthenticationType string
type matchExpressionType string

const (
	JwtType     AuthenticationType  = "JWT"
	ExactMatch  matchExpressionType = "exact"
	PrefixMatch matchExpressionType = "prefix"
	SuffixMatch matchExpressionType = "suffix"
	RegexMatch  matchExpressionType = "regex"
)

type TriggerRule struct {
	ExcludedPaths []MatchExpression `json:"excludedPaths,omitempty"`
}

type MatchExpression struct {
	ExprType matchExpressionType
	Value    string
}

func (self *MatchExpression) UnmarshalJSON(b []byte) error {

	var generic map[string]string
	if err := json.Unmarshal(b, &generic); err != nil {
		fmt.Println("Error during unmarshaling MatchExpression", err)
		return err
	}

	me, err := toMatchExpression(generic)
	if err != nil {
		return err
	}
	*self = *me

	return nil
}

func (self MatchExpression) MarshalJSON() ([]byte, error) {
	res := map[string]string{string(self.ExprType): self.Value}
	return json.Marshal(res)
}

func toMatchExpression(generic map[string]string) (*MatchExpression, error) {

	if len(generic) != 1 {
		return nil, errors.New(fmt.Sprintf("Expected exactly 1 entry in MatchExpression, got: %d", len(generic)))
	}

	for t, v := range generic {

		switch t {
		default:
			return nil, errors.New(fmt.Sprintf("Unknown MatchExpression type: \"%s\"", t))
		case string(ExactMatch):
			fallthrough
		case string(PrefixMatch):
			fallthrough
		case string(SuffixMatch):
			fallthrough
		case string(RegexMatch):
			return &MatchExpression{matchExpressionType(t), v}, nil
		}
	}

	return nil, errors.New("Failed to unmarshal MatchExpression.")
}

type JwtAuthentication struct {
	JwksUri     string       `json:"jwksUri"`
	Issuer      string       `json:"issuer"`
	TriggerRule *TriggerRule `json:"triggerRule,omitempty"`
}

type ApiStatus struct {
	ValidationStatus     kymaMeta.StatusCode            `json:"validationStatus,omitempty"`
	AuthenticationStatus kymaMeta.GatewayResourceStatus `json:"authenticationStatus,omitempty"`
	VirtualServiceStatus kymaMeta.GatewayResourceStatus `json:"virtualServiceStatus,omitempty"`
}

func (s *ApiStatus) IsEmpty() bool {
	return s.VirtualServiceStatus.IsEmpty() && s.AuthenticationStatus.IsEmpty() && s.ValidationStatus.IsEmpty()
}

func (s *ApiStatus) IsSuccessful() bool {
	return s.VirtualServiceStatus.IsSuccessful() && s.AuthenticationStatus.IsSuccessful() && s.ValidationStatus.IsSuccessful()
}

func (s *ApiStatus) IsInProgress() bool {
	return s.VirtualServiceStatus.IsInProgress() || s.AuthenticationStatus.IsInProgress() || s.ValidationStatus.IsInProgress()
}

func (s *ApiStatus) IsError() bool {
	return s.VirtualServiceStatus.IsError() || s.AuthenticationStatus.IsError() || s.ValidationStatus.IsError()
}

func (s *ApiStatus) IsHostnameOccupied() bool {
	return s.VirtualServiceStatus.IsHostnameOccupied()
}

func (s *ApiStatus) IsTargetServiceOccupied() bool {
	return s.ValidationStatus.IsTargetServiceOccupied()
}

func (s *ApiStatus) SetInProgress() {
	s.ValidationStatus = kymaMeta.InProgress
	s.AuthenticationStatus.Code = kymaMeta.InProgress
	s.VirtualServiceStatus.Code = kymaMeta.InProgress
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ApiList struct {
	k8sMeta.TypeMeta `json:",inline"`
	k8sMeta.ListMeta `json:"metadata,omitempty"`

	Items []Api `json:"items"`
}

// String method makes ApiSpec satisfy the Stringer interface and defines how the textual representation of the structure will look
func (s ApiSpec) String() string {
	disableIstioAuthPolicyMTLS := "<nil>"
	authenticationEnabled := "<nil>"
	authentication := "<nil slice>"

	if s.DisableIstioAuthPolicyMTLS != nil {
		disableIstioAuthPolicyMTLS = fmt.Sprintf("&%t", *s.DisableIstioAuthPolicyMTLS)
	}
	if s.AuthenticationEnabled != nil {
		authenticationEnabled = fmt.Sprintf("&%t", *s.AuthenticationEnabled)
	}
	if s.Authentication != nil {
		authentication = fmt.Sprintf("%+v", s.Authentication)
	}

	return fmt.Sprintf("{Service:%+v Hostname:%s DisableIstioAuthPolicyMTLS:%s AuthenticationEnabled:%s Authentication:%+v}", s.Service, s.Hostname, disableIstioAuthPolicyMTLS, authenticationEnabled, authentication)
}
