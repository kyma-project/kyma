package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/kyma-project/kyma/components/application-broker/internal"

	"github.com/pkg/errors"
)

var (
	whitespaces                  = regexp.MustCompile(`\s+`)
	dash                         = regexp.MustCompile(`-+`)
	envNameAllowedChars          = regexp.MustCompile(`[a-zA-Z_]+[a-zA-Z0-9_\s]*`)
	envNameSubstringAllowedChars = regexp.MustCompile(`[a-zA-Z0-9_]*`)
)

// BindingCredentialsRenderer provides functionality for rendering binding information
type BindingCredentialsRenderer struct {
	APIPackageCredGetter apiPackageCredentialsGetter
	GatewayBaseURLFormat string
	SBFetcher            ServiceBindingFetcher
}

// NewBindingCredentialsRenderer returns new instance of BindingCredentialsRenderer
func NewBindingCredentialsRenderer(APIPackageCredGetter apiPackageCredentialsGetter, gatewayBaseURLFormat string, fetcher ServiceBindingFetcher) *BindingCredentialsRenderer {
	return &BindingCredentialsRenderer{
		APIPackageCredGetter: APIPackageCredGetter,
		GatewayBaseURLFormat: gatewayBaseURLFormat,
		SBFetcher:            fetcher,
	}
}

// GetBindingCredentialsV2 returns binding information with API Package credential
func (b *BindingCredentialsRenderer) GetBindingCredentialsV2(ctx context.Context, ns string, service internal.Service, bindingID, appID, instanceID string) (map[string]interface{}, error) {
	pkgCreds, err := b.APIPackageCredGetter.GetAPIPackageCredentials(ctx, appID, string(service.ID), instanceID)
	if err != nil {
		return nil, errors.Wrap(err, "while getting API Package credentials")
	}

	jsonConfig, err := json.Marshal(pkgCreds.Config)
	if err != nil {
		return nil, errors.Wrap(err, "while marshaling API Package credentials")
	}

	secretName, err := b.SBFetcher.GetServiceBindingSecretName(ns, bindingID)
	if err != nil {
		return nil, errors.Wrapf(err, "while fetching secret name for service binding with ID: %q", bindingID)
	}

	creds := map[string]interface{}{
		"CREDENTIALS_TYPE": pkgCreds.Type,
		"CONFIGURATION":    jsonConfig,
	}
	for _, e := range service.Entries {
		if e.Type == internal.APIEntryType {
			creds[b.apiGatewayURLKey(e)] = b.apiGatewayURL(ns, secretName, e)
			creds[b.apiTargetURLKey(e)] = b.apiTargetURL(e)
		}
	}

	return creds, nil
}

func (b *BindingCredentialsRenderer) apiGatewayURLKey(e internal.Entry) string {
	return b.prefix(e) + "_GATEWAY_URL"
}

func (b *BindingCredentialsRenderer) apiGatewayURL(ns string, secretName string, e internal.Entry) string {
	baseURL := fmt.Sprintf(b.GatewayBaseURLFormat, ns)
	baseURL = strings.TrimSuffix(baseURL, "/")

	return fmt.Sprintf("%s/secret/%s/api/%s", baseURL, secretName, b.prefix(e))
}

func (b *BindingCredentialsRenderer) apiTargetURLKey(e internal.Entry) string {
	return b.prefix(e) + "_TARGET_URL"
}

func (b *BindingCredentialsRenderer) apiTargetURL(e internal.Entry) string {
	return e.TargetURL
}

// prefix returns a valid environment variable name prefix which consist of alphabetic characters, digits, '_' and does not start with a digit
func (b *BindingCredentialsRenderer) prefix(e internal.Entry) string {
	sanitizedName := b.sanitizeName(e.Name)
	sanitizedID := b.sanitizeID(e.ID)

	return strings.ToUpper(fmt.Sprintf("%s_%s", sanitizedName, sanitizedID))
}

func (b *BindingCredentialsRenderer) sanitizeName(in string) string {
	// remove not allowed characters like @,#,$ etc.
	in = strings.Join(envNameAllowedChars.FindAllString(in, -1), "")
	// remove leading and trailing white space
	in = strings.TrimSpace(in)
	// replace rest white space between words with underscore
	in = whitespaces.ReplaceAllString(in, "_")

	return in
}

func (b *BindingCredentialsRenderer) sanitizeID(in string) string {
	// replace dash in UUID with underscores
	in = dash.ReplaceAllString(in, "_")
	// ensure that not allowed characters are removed (just in case)
	in = strings.Join(envNameSubstringAllowedChars.FindAllString(in, -1), "")

	return in
}

// Deprecated, remove in https://github.com/kyma-project/kyma/issues/7415
// in old approach if it is bindable then it has only one API entry
func (b *BindingCredentialsRenderer) GetBindingCredentialsV1(_ context.Context, _ string, service internal.Service, _, _, _ string) (map[string]interface{}, error) {
	creds := make(map[string]interface{})
	creds[fieldNameGatewayURL] = service.Entries[0].GatewayURL
	return creds, nil
}
