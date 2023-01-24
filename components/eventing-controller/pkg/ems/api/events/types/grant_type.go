package types

type GrantType string

const GrantTypeClientCredentials GrantType = "client_credentials"

func IsInvalidGrantType(value string) bool {
	return value != string(GrantTypeClientCredentials)
}

func GetGrantType(_ string) GrantType {
	return GrantTypeClientCredentials
}
