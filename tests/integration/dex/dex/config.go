package dex

type idProviderConfig struct {
	dexConfig       dexConfig
	clientConfig    clientConfig
	userCredentials userCredentials
}

type dexConfig struct {
	baseUrl           string
	authorizeEndpoint string
	tokenEndpoint     string
}

type clientConfig struct {
	id          string
	redirectUri string
}

type userCredentials struct {
	username string
	password string
}
