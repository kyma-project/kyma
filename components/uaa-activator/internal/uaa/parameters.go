package uaa

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/pkg/errors"
)

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

func NewInstanceParameters(cfg Config, domain string) ([]byte, error) {
	parameters := Schema{
		Xsappname:  fmt.Sprintf("%s_%s", strings.ReplaceAll(domain, ".", "_"), randomString(5)),
		TenantMode: "shared",
		Scopes: []Scope{
			{
				Name:        "$XSAPPNAME.email",
				Description: "get user email",
			},
			{
				Name:        fmt.Sprintf("$XSAPPNAME.%s", cfg.DeveloperGroup),
				Description: "Runtime developer access to all managed resources",
			},
			{
				Name:        fmt.Sprintf("$XSAPPNAME.%s", cfg.NamespaceAdminGroup),
				Description: "Runtime admin access to all managed resources",
			},
		},
		Authorities: []string{
			"$ACCEPT_GRANTED_AUTHORITIES",
		},
		RoleTemplates: []RoleTemplate{
			{
				Name:        cfg.DeveloperRole,
				Description: "Runtime developer access to all managed resources",
				ScopeReferences: []string{
					fmt.Sprintf("$XSAPPNAME.%s", cfg.DeveloperGroup),
				},
			},
			{
				Name:        cfg.NamespaceAdminRole,
				Description: "Runtime admin access to all managed resources",
				ScopeReferences: []string{
					fmt.Sprintf("$XSAPPNAME.%s", cfg.NamespaceAdminGroup),
				},
			},
		},
		RoleCollections: []RoleCollection{
			{
				Name:        cfg.DeveloperRole,
				Description: "Kyma Runtime Developer Role Collection for development tasks in given custom namespaces",
				RoleTemplateReference: []string{
					fmt.Sprintf("$XSAPPNAME.%s", cfg.DeveloperRole),
				},
			},
			{
				Name:        cfg.NamespaceAdminRole,
				Description: "Kyma Runtime Namespace Admin Role Collection for administration tasks across all custom namespaces",
				RoleTemplateReference: []string{
					fmt.Sprintf("$XSAPPNAME.%s", cfg.NamespaceAdminRole),
				},
			},
		},
		Oauth2Configuration: Oauth2Configuration{
			RedirectUris: []string{
				fmt.Sprintf("https://dex.%s/callback", strings.Trim(domain, "/")),
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

func randomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
