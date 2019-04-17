package k8s

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
	v1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type selfSubjectRulesResolver struct {
	gqlSelfSubjectRulesConverter
	selfSubjectRulesSvc
}

//go:generate mockery -name=selfSubjectRulesSvc -output=automock -outpkg=automock -case=underscore
type selfSubjectRulesSvc interface {
	Create(ssrr *v1.SelfSubjectRulesReview) (*v1.SelfSubjectRulesReview, error)
}

//go:generate mockery -name=gqlSelfSubjectRulesConverter -output=automock -outpkg=automock -case=underscore
type gqlSelfSubjectRulesConverter interface {
	ToGQL(in *v1.SelfSubjectRulesReview) (*gqlschema.SelfSubjectRules, error)
}

func newSelfSubjectRulesResolver(selfSubjectRulesSvc selfSubjectRulesSvc) *selfSubjectRulesResolver {
	return &selfSubjectRulesResolver{
		selfSubjectRulesSvc:          selfSubjectRulesSvc,
		gqlSelfSubjectRulesConverter: &selfSubjectRulesConverter{},
	}
}

func (r *selfSubjectRulesResolver) SelfSubjectRulesQuery(ctx context.Context, namespace *string) (*gqlschema.SelfSubjectRules, error) {
	ssrrin := &v1.SelfSubjectRulesReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SelfSubjectRulesReview",
			APIVersion: "authorization.k8s.io/v1",
		},
		Spec: v1.SelfSubjectRulesReviewSpec{
			Namespace: *namespace,
		},
	}
	ssrrout, err := r.selfSubjectRulesSvc.Create(ssrrin)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while reviewing self subject rules"))
		return &gqlschema.SelfSubjectRules{}, gqlerror.New(err, pretty.SelfSubjectRules)
	}
	return r.gqlSelfSubjectRulesConverter.ToGQL(ssrrout)
}
