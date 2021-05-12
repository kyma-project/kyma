package k8s

import (
	"context"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/authn"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	authv1 "k8s.io/api/authorization/v1"
	v1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
)

type selfSubjectRulesService struct {
	client v1.AuthorizationV1Interface
}

func newSelfSubjectRulesService(client v1.AuthorizationV1Interface) *selfSubjectRulesService {
	return &selfSubjectRulesService{
		client: client,
	}
}

func (svc *selfSubjectRulesService) Create(ctx context.Context, ssrr []byte) (result *authv1.SelfSubjectRulesReview, err error) {
	if ssrr == nil {
		err := gqlerror.New(err, pretty.SelfSubjectRules)
		return &authv1.SelfSubjectRulesReview{}, err
	}
	u, err := authn.UserInfoForContext(ctx)
	username := u.GetName()
	result = &authv1.SelfSubjectRulesReview{}
	err = svc.client.RESTClient().Post().
		AbsPath("/apis/authorization.k8s.io/v1").
		Resource("selfsubjectrulesreviews").
		SetHeader("Impersonate-User", username).
		SetHeader("Impersonate-Group", u.GetGroups()...).
		Body(ssrr).
		Do(context.Background()).
		Into(result)
	return
}
