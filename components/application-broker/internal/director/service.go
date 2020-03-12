package director

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kyma-project/kyma/components/application-broker/internal"

	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/authorization"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/proxyconfig"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
)

const contextNameForInstanceID = "instance_id"

// APIPackageInstanceAuthClient defines required Director client functionality to handle the
// package instance auth creation and deletion
type APIPackageInstanceAuthClient interface {
	RequestPackageInstanceAuth(context.Context, RequestPackageInstanceAuthInput) (*RequestPackageInstanceAuthOutput, error)
	FindPackageInstanceAuths(context.Context, GetPackageInstanceAuthsInput) (*GetPackageInstanceAuthsOutput, error)
	FindPackageInstanceAuth(context.Context, GetPackageInstanceAuthInput) (*GetPackageInstanceAuthOutput, error)
	RequestPackageInstanceAuthDeletion(context.Context, RequestPackageInstanceAuthDeletionInput) (*RequestPackageInstanceAuthDeletionOutput, error)
}

// Service is a wrapper around the Director GraphQL client to support business logic
type Service struct {
	directorCli              APIPackageInstanceAuthClient
	operationPollingInterval time.Duration
	operationPollingTimeout  time.Duration
}

// NewService returns new Service instance
func NewService(cli APIPackageInstanceAuthClient, cfg ServiceConfig) *Service {
	return &Service{
		directorCli:              cli,
		operationPollingInterval: cfg.OperationPollingInterval,
		operationPollingTimeout:  cfg.OperationPollingTimeout,
	}
}

// EnsureAPIPackageCredentials ensures that the API Package credentials were created for the given service instance ID
func (s *Service) EnsureAPIPackageCredentials(ctx context.Context, appID, pkgID, instanceID string, inputSchema map[string]interface{}) error {
	timeout, cancel := context.WithTimeout(ctx, s.operationPollingTimeout)
	defer cancel()
	auth, err := s.getOrCreatePackageInstanceAuth(timeout, appID, pkgID, instanceID, inputSchema)
	if err != nil {
		return errors.Wrap(err, "while getting or creating package instance auth")
	}

	switch {
	case s.isSucceeded(auth.Status):
		return nil
	case s.isFailed(auth.Status):
		return errors.Errorf("requested package instance auth failed, got status %+v", *auth.Status)
	default:
		return s.waitForPackageInstanceAuth(timeout, appID, pkgID, auth.ID)
	}
}

// GetAPIPackageCredentials returns API Package credentials associated with the given service instance id
func (s *Service) GetAPIPackageCredentials(ctx context.Context, appID, pkgID, instanceID string) (internal.APIPackageCredential, error) {
	timeout, cancel := context.WithTimeout(ctx, s.operationPollingTimeout)
	defer cancel()

	auth, err := s.findPackageInstanceAuthByInstanceID(timeout, appID, pkgID, instanceID)
	if err != nil {
		return internal.APIPackageCredential{}, errors.Wrapf(err, "while finding package instance auth for instance id %q", instanceID)
	}

	if auth == nil {
		return internal.APIPackageCredential{}, errors.Errorf("package instance auth not found for instance id %q", instanceID)
	}

	if auth.Status == nil {
		return internal.APIPackageCredential{}, errors.New("requested package instance auth has nil status")
	}
	if auth.Status.Condition != schema.PackageInstanceAuthStatusConditionSucceeded {
		return internal.APIPackageCredential{}, errors.Errorf("requested package instance auth is not in succeeded state, got status [%+v]", auth.Status)
	}
	if auth.Auth == nil {
		return internal.APIPackageCredential{}, errors.Errorf("package instance auth %q is in success state but has an empty Auth", auth.ID)
	}

	model, err := s.mapPackageInstanceAuthToModel(*auth)
	if err != nil {
		return internal.APIPackageCredential{}, errors.Wrap(err, "while mapping PackageInstanceAuth dto to model")
	}

	return model, nil
}

// EnsureAPIPackageCredentialsDeleted ensures that the given API Package credential associated with the given service instance id
// is removed.
func (s *Service) EnsureAPIPackageCredentialsDeleted(ctx context.Context, appID string, pkgID string, instanceID string) error {
	timeout, cancel := context.WithTimeout(ctx, s.operationPollingTimeout)
	defer cancel()

	auth, err := s.findPackageInstanceAuthByInstanceID(timeout, appID, pkgID, instanceID)
	if err != nil {
		return errors.Wrap(err, "while getting package instance auth for deletion process")
	}
	if auth == nil {
		return nil
	}

	_, err = s.directorCli.RequestPackageInstanceAuthDeletion(timeout, RequestPackageInstanceAuthDeletionInput{
		InstanceAuthID: auth.ID,
	})
	switch {
	case err == nil:
	case IsGQLNotFoundError(err):
		return nil
	default:
		return errors.Wrapf(err, "while calling Director to delete package %q instance auth", pkgID)
	}

	return s.waitForDeletedPackageInstanceAuth(timeout, appID, pkgID, auth.ID)
}

func (s *Service) mapPackageInstanceAuthToModel(pkgAuth schema.PackageInstanceAuth) (internal.APIPackageCredential, error) {
	var (
		auth = pkgAuth.Auth
		cfg  = proxyconfig.Configuration{}
	)

	if auth.RequestAuth != nil && auth.RequestAuth.Csrf != nil {
		cfg.CSRFConfig = &proxyconfig.CSRFConfig{TokenURL: auth.RequestAuth.Csrf.TokenEndpointURL}
	}

	if auth.AdditionalHeaders != nil {
		s.initIfNil(cfg.RequestParameters)
		cfg.RequestParameters.Headers = (*map[string][]string)(auth.AdditionalHeaders)
	}

	if auth.AdditionalQueryParams != nil {
		s.initIfNil(cfg.RequestParameters)
		cfg.RequestParameters.QueryParameters = (*map[string][]string)(auth.AdditionalQueryParams)
	}

	var credType proxyconfig.AuthType
	switch c := auth.Credential.(type) {
	case nil:
		credType = proxyconfig.NoAuth
	case *schema.OAuthCredentialData:
		credType = proxyconfig.Oauth
		cfg.Credentials = proxyconfig.OauthConfig{
			ClientId:     c.ClientID,
			ClientSecret: c.ClientSecret,
			TokenURL:     c.URL,
		}
	case *schema.BasicCredentialData:
		credType = proxyconfig.Basic
		cfg.Credentials = proxyconfig.BasicAuthConfig{
			Username: c.Username,
			Password: c.Password,
		}
	default:
		return internal.APIPackageCredential{}, errors.Errorf("got unknown credential type %T", c)
	}

	return internal.APIPackageCredential{
		ID:     pkgAuth.ID,
		Type:   credType,
		Config: cfg,
	}, nil
}

func (s *Service) initIfNil(req *authorization.RequestParameters) {
	if req == nil {
		req = &authorization.RequestParameters{}
	}
}

func (s *Service) isAuthAttachedForInstance(auth schema.PackageInstanceAuth, instanceID string) (bool, error) {
	if auth.Context == nil {
		return false, nil
	}

	var authContext map[string]string
	if err := json.Unmarshal([]byte(*auth.Context), &authContext); err != nil {
		return false, errors.Wrap(err, "while unmarshaling auth context")
	}

	if authContext[contextNameForInstanceID] == instanceID {
		return true, nil
	}

	return false, nil
}

func (s *Service) waitForDeletedPackageInstanceAuth(ctx context.Context, appID, pkgID, authID string) error {
	isPackageInstanceAuthDeleted := func() (done bool, err error) {
		out, err := s.directorCli.FindPackageInstanceAuth(ctx, GetPackageInstanceAuthInput{
			PackageID:      pkgID,
			ApplicationID:  appID,
			InstanceAuthID: authID,
		})
		if err != nil {
			return false, errors.Wrapf(err, "while trying to get fresh package %q instance auth", pkgID)
		}
		if out.InstanceAuth == nil {
			return true, nil
		}
		return false, nil
	}

	err := wait.PollUntil(s.operationPollingInterval, isPackageInstanceAuthDeleted, ctx.Done())
	if err != nil {
		return errors.Wrap(err, "while waiting for deleted package instance auth")
	}

	return nil
}

func (s *Service) waitForPackageInstanceAuth(ctx context.Context, appID, pkgID, authID string) error {
	isPkgInstanceAuthCreated := func() (done bool, err error) {
		out, err := s.directorCli.FindPackageInstanceAuth(ctx, GetPackageInstanceAuthInput{
			PackageID:      pkgID,
			ApplicationID:  appID,
			InstanceAuthID: authID,
		})
		if err != nil {
			return false, errors.Wrap(err, "while trying to get fresh status of instance auth")
		}

		switch {
		case out.InstanceAuth == nil:
			return false, nil
		case s.isSucceeded(out.InstanceAuth.Status):
			return true, nil
		case s.isFailed(out.InstanceAuth.Status):
			return false, errors.Errorf("requesting package instance auth failed, got status [%+v]", out.InstanceAuth.Status)
		}
		return false, nil
	}

	err := wait.PollUntil(s.operationPollingInterval, isPkgInstanceAuthCreated, ctx.Done())
	if err != nil {
		return errors.Wrap(err, "while waiting for package instance auth")
	}

	return nil
}
func (s *Service) getOrCreatePackageInstanceAuth(ctx context.Context, appID, pkgID, instanceID string, inputSchema map[string]interface{}) (schema.PackageInstanceAuth, error) {
	auth, err := s.findPackageInstanceAuthByInstanceID(ctx, appID, pkgID, instanceID)
	if err != nil {
		return schema.PackageInstanceAuth{}, err
	}
	if auth != nil {
		return *auth, nil
	}

	// auth not found, so create a new one
	out, err := s.directorCli.RequestPackageInstanceAuth(ctx, RequestPackageInstanceAuthInput{
		PackageID:   pkgID,
		Context:     Values{contextNameForInstanceID: instanceID},
		InputSchema: inputSchema,
	})
	if err != nil {
		return schema.PackageInstanceAuth{}, errors.Wrapf(err, "while calling Director to request package %q instance auth", pkgID)
	}
	return out.InstanceAuth, nil
}

func (s *Service) findPackageInstanceAuthByInstanceID(ctx context.Context, appID, pkgID, instanceID string) (*schema.PackageInstanceAuth, error) {
	out, err := s.directorCli.FindPackageInstanceAuths(ctx, GetPackageInstanceAuthsInput{
		ApplicationID: appID,
		PackageID:     pkgID,
	})
	if err != nil {
		return nil, err
	}

	for _, auth := range out.InstanceAuths {
		if auth == nil {
			continue
		}

		attached, err := s.isAuthAttachedForInstance(*auth, instanceID)
		if err != nil {
			return nil, err
		}
		if attached {
			return auth, nil
		}
	}

	return nil, nil
}

func (*Service) isSucceeded(status *schema.PackageInstanceAuthStatus) bool {
	if status == nil {
		return false
	}
	if status.Condition == schema.PackageInstanceAuthStatusConditionSucceeded {
		return true
	}
	return false
}

func (*Service) isFailed(status *schema.PackageInstanceAuthStatus) bool {
	if status == nil {
		return false
	}
	if status.Condition == schema.PackageInstanceAuthStatusConditionFailed {
		return true
	}
	return false
}
