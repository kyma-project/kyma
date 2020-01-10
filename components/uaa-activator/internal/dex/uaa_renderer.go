package dex

import (
	"bytes"
	"context"
	"text/template"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const uaaConnector = `- type: xsuaa
  id: xsuaa
  name: XSUAA
  config:
    issuer: "{{ .URL | toString }}/oauth/token"
    clientID: "{{ .ClientID | toString }}"
    clientSecret: "{{ .ClientSecret | toString }}"
    redirectURI: "https://dex.{{ .ClusterDomain }}/callback"
    userNameKey: "user_name"
    appname: "{{ .Xsappname | toString }}"`

type (
	UAASecret struct {
		ClientID     []byte `mapstructure:"clientid"`
		ClientSecret []byte `mapstructure:"clientsecret"`
		Xsappname    []byte `mapstructure:"xsappname"`
		URL          []byte `mapstructure:"url"`
	}

	UAATmplData struct {
		UAASecret
		ClusterDomain string
	}

	uaaConfigProvider interface {
		RenderUAAConnectorConfig(ctx context.Context) (string, error)
	}
)

// UAARenderer provides functionality for renderng the UAA connector config suited for Dex
type UAARenderer struct {
	cli           client.Client
	secret        client.ObjectKey
	clusterDomain string
}

// NewUAARenderer returns a new UAARenderer instance
func NewUAARenderer(cli client.Client, secret client.ObjectKey, clusterDomain string) *UAARenderer {
	return &UAARenderer{
		cli:           cli,
		secret:        secret,
		clusterDomain: clusterDomain,
	}
}

// RenderUAAConnectorConfig renders uaa connector configuration based on information stored in Secret.
func (u *UAARenderer) RenderUAAConnectorConfig(ctx context.Context) (string, error) {
	uaaConnectorTmpl, err := template.New("uaa-connector").Funcs(template.FuncMap{
		"toString": func(feature []byte) string {
			return string(feature)
		},
	}).Parse(uaaConnector)
	if err != nil {
		return "", errors.Wrap(err, "while parsing uaa connector template")
	}

	rawSecret := v1.Secret{}
	if err := u.cli.Get(ctx, u.secret, &rawSecret); err != nil {
		return "", errors.Wrapf(err, "while fetching Secret %v", u.secret.String())
	}

	uaaSecret := UAASecret{}
	if err := mapstructure.Decode(rawSecret.Data, &uaaSecret); err != nil {
		return "", errors.Wrap(err, "while mapping Secret into typed struct")
	}

	tmplData := UAATmplData{
		UAASecret:     uaaSecret,
		ClusterDomain: u.clusterDomain,
	}

	renderedConnector := &bytes.Buffer{}
	if err := uaaConnectorTmpl.Execute(renderedConnector, tmplData); err != nil {
		return "", errors.Wrapf(err, "while executing uaa connector template")
	}

	return renderedConnector.String(), nil
}
