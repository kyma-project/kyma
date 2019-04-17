package k8s

import (
	"context"

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

func (svc *selfSubjectRulesService) Create(ctx context.Context, ssrr *authv1.SelfSubjectRulesReview) (result *authv1.SelfSubjectRulesReview, err error) {

	username := ctx.Value("username").(string)
	result = &authv1.SelfSubjectRulesReview{}
	err = svc.client.RESTClient().Post().
		AbsPath("/apis/authorization.k8s.io/v1").
		Resource("selfsubjectrulesreviews").
		SetHeader("Impersonate-User", username).
		Body(ssrr).
		Do().
		Into(result)
	return
}
