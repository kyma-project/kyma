/*
 *  © 2018 SAP SE or an SAP affiliate company.
 *  All rights reserved.
 *  Please see http://www.sap.com/corporate-en/legal/copyright/index.epx for additional trademark information and
 *  notices.
 */
package testkit

import (
	"bytes"
	"encoding/json"
)

var (
	ApiRawSpec    = compact([]byte("{\"name\":\"api\"}"))
	EventsRawSpec = compact([]byte("{\"name\":\"events\"}"))
)

type Service struct {
	ID          string `json:"id"`
	Provider    string `json:"provider"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ServiceDetails struct {
	Provider      string         `json:"provider"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Api           *API           `json:"api,omitempty"`
	Events        *Events        `json:"events,omitempty"`
	Documentation *Documentation `json:"documentation,omitempty"`
}

type PostServiceResponse struct {
	ID string `json:"id"`
}

type ErrorResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

type API struct {
	TargetUrl   string          `json:"targetUrl"`
	Credentials *Credentials    `json:"credentials,omitempty"`
	Spec        json.RawMessage `json:"spec,omitempty"`
}

type Credentials struct {
	Oauth Oauth `json:"oauth"`
}

type Oauth struct {
	URL          string `json:"url"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

type Events struct {
	Spec json.RawMessage `json:"spec,omitempty"`
}

type Documentation struct {
	DisplayName string       `json:"displayName"`
	Description string       `json:"description"`
	Type        string       `json:"type"`
	Tags        []string     `json:"tags,omitempty"`
	Docs        []DocsObject `json:"docs,omitempty"`
}

type DocsObject struct {
	Title  string `json:"title"`
	Type   string `json:"type"`
	Source string `json:"source"`
}

func compact(src []byte) []byte {
	buffer := new(bytes.Buffer)
	err := json.Compact(buffer, src)
	if err != nil {
		return src
	}
	return buffer.Bytes()
}
