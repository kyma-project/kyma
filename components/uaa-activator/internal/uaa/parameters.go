package uaa

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/pkg/errors"
)

type ParametersBuilder struct {
	domain        string
	redirectURL   string
	developerRole string
	adminRole     string
	config        Config
}

func NewParametersBuilder(cfg Config, domain string) *ParametersBuilder {
	shootName := strings.Split(domain, ".")[0]
	return &ParametersBuilder{
		domain:        domain,
		redirectURL:   fmt.Sprintf("https://dex.%s/callback", strings.Trim(domain, "/")),
		developerRole: roleName(cfg.DeveloperRole, shootName),
		adminRole:     roleName(cfg.NamespaceAdminRole, shootName),
		config:        cfg,
	}
}

type Schema struct {
	Xsappname           string              `json:"xsappname"`
	TenantMode          string              `json:"tenant-mode"`
	Scopes              []Scope             `json:"scopes"`
	Authorities         []string            `json:"authorities"`
	RoleTemplates       []RoleTemplate      `json:"role-templates"`
	RoleCollections     []RoleCollection    `json:"role-collections"`
	Oauth2Configuration Oauth2Configuration `json:"oauth2-configuration"`
}

type Scope struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type RoleTemplate struct {
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	ScopeReferences []string `json:"scope-references"`
}

type RoleCollection struct {
	Name                  string   `json:"name"`
	Description           string   `json:"description"`
	RoleTemplateReference []string `json:"role-template-references"`
}

type Oauth2Configuration struct {
	RedirectUris     []string `json:"redirect-uris"`
	SystemAttributes []string `json:"system-attributes"`
}

func (pb *ParametersBuilder) Generate(instance *v1beta1.ServiceInstance) ([]byte, error) {
	xsappname, err := pb.xsappname(instance)
	if err != nil {
		return []byte{}, errors.Wrap(err, "while generating xsappname")
	}
	parameters := Schema{
		Xsappname:  xsappname,
		TenantMode: "shared",
		Scopes: []Scope{
			{
				Name:        "$XSAPPNAME.email",
				Description: "get user email",
			},
			{
				Name:        fmt.Sprintf("$XSAPPNAME.%s", pb.config.DeveloperGroup),
				Description: "Runtime developer access to all managed resources",
			},
			{
				Name:        fmt.Sprintf("$XSAPPNAME.%s", pb.config.NamespaceAdminGroup),
				Description: "Runtime admin access to all managed resources",
			},
		},
		Authorities: []string{
			"$ACCEPT_GRANTED_AUTHORITIES",
		},
		RoleTemplates: []RoleTemplate{
			{
				Name:        pb.config.DeveloperRole,
				Description: "Runtime developer access to all managed resources",
				ScopeReferences: []string{
					fmt.Sprintf("$XSAPPNAME.%s", pb.config.DeveloperGroup),
				},
			},
			{
				Name:        pb.config.NamespaceAdminRole,
				Description: "Runtime admin access to all managed resources",
				ScopeReferences: []string{
					fmt.Sprintf("$XSAPPNAME.%s", pb.config.NamespaceAdminGroup),
				},
			},
		},
		RoleCollections: []RoleCollection{
			{
				Name:        pb.developerRole,
				Description: "Kyma Runtime Developer Role Collection for development tasks in given custom namespaces",
				RoleTemplateReference: []string{
					fmt.Sprintf("$XSAPPNAME.%s", pb.config.DeveloperRole),
				},
			},
			{
				Name:        pb.adminRole,
				Description: "Kyma Runtime Namespace Admin Role Collection for administration tasks across all custom namespaces",
				RoleTemplateReference: []string{
					fmt.Sprintf("$XSAPPNAME.%s", pb.config.NamespaceAdminRole),
				},
			},
		},
		Oauth2Configuration: Oauth2Configuration{
			RedirectUris: []string{
				pb.redirectURL,
			},
			SystemAttributes: []string{
				"groups",
				"rolecollections",
			},
		},
	}

	marshalledParams, err := json.MarshalIndent(parameters, "", "  ")
	if err != nil {
		return []byte{}, errors.Wrap(err, "while marshaling parameters")
	}

	return marshalledParams, nil
}

func (pb *ParametersBuilder) xsappname(instance *v1beta1.ServiceInstance) (string, error) {
	if instance == nil {
		return fmt.Sprintf("%s_%s", strings.ReplaceAll(pb.domain, ".", "_"), randomString(5)), nil
	}
	if instance.Spec.ParametersFrom != nil {
		// if ParametersFrom are not nil, it means ServiceInstance comes from SKR before 1.17
		// which means xsappname was just a domain name with replaced chars
		return strings.ReplaceAll(pb.domain, ".", "_"), nil
	}
	schema := &Schema{}
	err := json.Unmarshal(instance.Spec.Parameters.Raw, schema)
	if err != nil {
		return "", errors.Wrap(err, "while unmarshal ServiceInstance parameters")
	}
	return schema.Xsappname, nil
}

func randomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// roleName creates proper name for RoleTemplates and RoleCollections
// according to SM the name may only include characters 'a'-'z', 'A'-'Z', '0'-'9', and '_'
func roleName(name, domain string) string {
	r := strings.NewReplacer(".", "_", ",", "_", ":", "", ";", "", "-", "_", "/", "", "\\", "")
	return fmt.Sprintf("%s__%s", r.Replace(name), r.Replace(domain))
}
