package migrator

import (
	kubelessv1beta1 "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kyma-project/kyma/components/function-controller/migrator/pkg/resource/apirule"
	apiruleTypes "github.com/kyma-project/kyma/components/function-controller/migrator/pkg/resource/apirule/types"
	"github.com/kyma-project/kyma/components/function-controller/migrator/pkg/resource/apis"
	apiTypes "github.com/kyma-project/kyma/components/function-controller/migrator/pkg/resource/apis/types"

	"github.com/pkg/errors"
)

type ApiOperator struct {
	Data   *apiTypes.Api
	ResCli apis.Api
}

func (m migrator) createApirules() error {
	for _, item := range m.apis {
		cli := apirule.New(m.dynamicCli, item.Data.Name, item.Data.Namespace, m.cfg.WaitTimeout, m.log.Info)
		m.log.WithValues("Name", item.Data.Name,
			"Namespace", item.Data.Namespace,
			"GroupVersion", apiruleTypes.GroupVersion.String(),
		).Info("Creating APIRule")

		if err := cli.Create(*item.Data, m.cfg.Domain); err != nil {
			return errors.Wrapf(err, "while creating ApiRule %s in namespace %s", item.Data.Name, item.Data.Namespace)
		}
	}

	return nil
}

func (m migrator) deleteApis() error {
	for _, item := range m.apis {
		m.log.WithValues("name", item.Data.Name,
			"namespace", item.Data.Namespace,
			"GroupVersion", kubelessv1beta1.SchemeGroupVersion.String(),
		).Info("Deleting api.gateway.kyma-project.io")

		if err := item.ResCli.Delete(); err != nil {
			return errors.Wrapf(err, "while deleting api.gateway.kyma-project.io %s in namespace %s", item.Data.Name, item.Data.Namespace)
		}
	}
	return nil
}
