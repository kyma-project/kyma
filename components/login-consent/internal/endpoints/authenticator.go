package endpoints

import (
	"context"

	"github.com/coreos/go-oidc"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type Authenticator struct {
	tokenSupport TokenSupport
	ctx          context.Context
}

func NewAuthenticator(dexAddress string, clientID string, clientSecret string, redirectURL string, scopes []string) (*Authenticator, error) {
	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, dexAddress)
	if err != nil {
		log.Info(err)
		return nil, err
	}

	oauth2Config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  redirectURL,
		Scopes:       scopes,
	}

	tsi := TokenSupportImpl{
		clientConfig: oauth2Config,
		provider:     provider,
	}

	return &Authenticator{
		tokenSupport: &tsi,
		ctx:          ctx,
	}, nil
}

type TokenSupport interface {
	//delegates to oauth3.Config.AuthCodeURL()
	AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
	//delegates to oauth2.Config.Exchange()
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	//delegates to oauth2.Config.ClientID
	ClientID() string
	//delegates to oidc.Provider.Verifier().Verify()
	Verify(config *oidc.Config, ctx context.Context, rawIDToken string) (*oidc.IDToken, error)
}

type TokenSupportImpl struct {
	clientConfig oauth2.Config
	provider     *oidc.Provider
}

func (tsi *TokenSupportImpl) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return tsi.clientConfig.AuthCodeURL(state, opts...)
}

func (tsi *TokenSupportImpl) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return tsi.clientConfig.Exchange(ctx, code, opts...)
}

func (tsi *TokenSupportImpl) ClientID() string {
	return tsi.clientConfig.ClientID //TODO: Is it OK? Old: TODO provide proper data here.
}

func (tsi *TokenSupportImpl) Verify(config *oidc.Config, ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	return tsi.provider.Verifier(config).Verify(ctx, rawIDToken)
}
