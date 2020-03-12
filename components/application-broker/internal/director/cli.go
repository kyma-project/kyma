package director

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/asaskevich/govalidator"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli"
	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

type GQLClient struct {
	cli gqlcli.GraphQLClient
}

func NewQGLClient(cli gqlcli.GraphQLClient) *GQLClient {
	return &GQLClient{cli: cli}
}

type RequestPackageInstanceAuthInput struct {
	PackageID   string `valid:"required"`
	Context     Values
	InputSchema Values
}

type RequestPackageInstanceAuthOutput struct {
	InstanceAuth schema.PackageInstanceAuth `json:"result"`
}

func (r *GQLClient) RequestPackageInstanceAuth(ctx context.Context, in RequestPackageInstanceAuthInput) (*RequestPackageInstanceAuthOutput, error) {
	if _, err := govalidator.ValidateStruct(in); err != nil {
		return nil, errors.Wrap(err, "while validating input")
	}

	input, err := in.InputSchema.MarshalToQGLJSON()
	if err != nil {
		return nil, errors.Wrap(err, "while marshaling input schema to GQL JSON")
	}

	inContext, err := in.Context.MarshalToQGLJSON()
	if err != nil {
		return nil, errors.Wrap(err, "while marshaling context to GQL JSON")
	}

	gqlRequest := gcli.NewRequest(fmt.Sprintf(`mutation {
			  result: requestPackageInstanceAuthCreation(
				packageID: "%s"
				in: {
				  context: %s
    			  inputParams: %s
				}
			  ) {
					id
					context
					auth {
					  additionalHeaders
					  additionalQueryParams
					  requestAuth {
						csrf {
						  tokenEndpointURL
						}
					  }
					  credential {
						... on OAuthCredentialData {
						  clientId
						  clientSecret
						  url
						}
						... on BasicCredentialData {
						  username
						  password
						}
					  }
					}
					status {
					  condition
					  timestamp
					  message
					  reason
					}
			  	 }
				}`, in.PackageID, inContext, input))

	var resp RequestPackageInstanceAuthOutput
	if err = r.cli.Run(ctx, gqlRequest, &resp); err != nil {
		return nil, errors.Wrap(err, "while executing GraphQL call to create package instance auth")
	}

	return &resp, nil
}

type GetPackageInstanceAuthsInput struct {
	ApplicationID string `valid:"required"`
	PackageID     string `valid:"required"`
}

type GetPackageInstanceAuthsOutput struct {
	InstanceAuths []*schema.PackageInstanceAuth
}

func (r *GQLClient) FindPackageInstanceAuths(ctx context.Context, in GetPackageInstanceAuthsInput) (*GetPackageInstanceAuthsOutput, error) {
	if _, err := govalidator.ValidateStruct(in); err != nil {
		return nil, errors.Wrap(err, "while validating input")
	}

	gqlRequest := gcli.NewRequest(fmt.Sprintf(`query {
			result: application(id: "%s") {
						package(id: "%s") {
						  instanceAuths {
							id
							context
							auth {
							  additionalHeaders
							  additionalQueryParams
							  requestAuth {
								csrf {
								  tokenEndpointURL
								}
							  }
							  credential {
								... on OAuthCredentialData {
								  clientId
								  clientSecret
								  url
								}
								... on BasicCredentialData {
								  username
								  password
								}
							  }
							}
							status {
							  condition
							  timestamp
							  message
							  reason
							}
						  }
						}
					  }
					}`, in.ApplicationID, in.PackageID))

	var resp struct {
		Result schema.ApplicationExt `json:"result"`
	}
	err := r.cli.Run(ctx, gqlRequest, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "while executing GraphQL call to get package instance auths")
	}

	return &GetPackageInstanceAuthsOutput{
		InstanceAuths: resp.Result.Package.InstanceAuths,
	}, nil
}

type GetPackageInstanceAuthInput struct {
	PackageID      string `valid:"required"`
	ApplicationID  string `valid:"required"`
	InstanceAuthID string `valid:"required"`
}

type GetPackageInstanceAuthOutput struct {
	InstanceAuth *schema.PackageInstanceAuth `json:"result"`
}

func (r *GQLClient) FindPackageInstanceAuth(ctx context.Context, in GetPackageInstanceAuthInput) (*GetPackageInstanceAuthOutput, error) {
	if _, err := govalidator.ValidateStruct(in); err != nil {
		return nil, errors.Wrap(err, "while validating input")
	}

	gqlRequest := gcli.NewRequest(fmt.Sprintf(`query {
			  result: application(id: %q) {
						package(id: %q) {
						  instanceAuth(id: %q) {
							id
							auth {
							  additionalHeaders
							  additionalQueryParams
							  requestAuth {
								csrf {
								  tokenEndpointURL
								}
							  }
							  credential {
								... on OAuthCredentialData {
								  clientId
								  clientSecret
								  url
								}
								... on BasicCredentialData {
								  username
								  password
								}
							  }
							}
							status {
							  condition
							  timestamp
							  message
							  reason
							}
						  }
						}
					  }
					}`, in.ApplicationID, in.PackageID, in.InstanceAuthID))

	var response struct {
		Result schema.ApplicationExt `json:"result"`
	}
	if err := r.cli.Run(ctx, gqlRequest, &response); err != nil {
		return nil, errors.Wrap(err, "while executing GraphQL call to get package instance auth")
	}

	return &GetPackageInstanceAuthOutput{
		InstanceAuth: response.Result.Package.InstanceAuth,
	}, nil
}

type RequestPackageInstanceAuthDeletionInput struct {
	InstanceAuthID string `valid:"required"`
}

type RequestPackageInstanceAuthDeletionOutput struct {
	ID     string                           `json:"id"`
	Status schema.PackageInstanceAuthStatus `json:"status"`
}

func (r *GQLClient) RequestPackageInstanceAuthDeletion(ctx context.Context, in RequestPackageInstanceAuthDeletionInput) (*RequestPackageInstanceAuthDeletionOutput, error) {
	if _, err := govalidator.ValidateStruct(in); err != nil {
		return nil, errors.Wrap(err, "while validating input")
	}

	gqlRequest := gcli.NewRequest(fmt.Sprintf(`mutation {
			  result: requestPackageInstanceAuthDeletion(authID: %q) {
						id
						status {
						  condition
						  timestamp
						  message
						  reason
						}
					  }
					}`, in.InstanceAuthID))

	var resp struct {
		Result RequestPackageInstanceAuthDeletionOutput `json:"result"`
	}
	if err := r.cli.Run(ctx, gqlRequest, &resp); err != nil {
		return nil, errors.Wrap(err, "while executing GraphQL call to delete the package instance auth")
	}

	return &resp.Result, nil
}

type Values map[string]interface{}

func (r *Values) MarshalToQGLJSON() (string, error) {
	input, err := json.Marshal(r)
	if err != nil {
		return "", err
	}

	return strconv.Quote(string(input)), nil
}
