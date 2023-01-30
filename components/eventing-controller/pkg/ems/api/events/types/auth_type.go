package types

type AuthType string

const (
	AuthTypeClientCredentials AuthType = "oauth2"
)

func IsInvalidAuthType(value string) bool {
	return value != string(AuthTypeClientCredentials)
}

func GetAuthType(_ string) AuthType {
	return AuthTypeClientCredentials
}
