package types

type AuthType string

const (
	AuthTypeBasic             AuthType = "basic"
	AuthTypeClientCredentials AuthType = "oauth2"
)
