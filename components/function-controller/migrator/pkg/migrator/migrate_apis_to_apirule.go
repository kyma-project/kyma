package migrator

import (
	"fmt"

	kubelessv1beta1 "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kyma-project/kyma/components/function-controller/migrator/pkg/resource/apirule"
	apiruleTypes "github.com/kyma-project/kyma/components/function-controller/migrator/pkg/resource/apirule/types"
	"github.com/kyma-project/kyma/components/function-controller/migrator/pkg/resource/apis"
	apiTypes "github.com/kyma-project/kyma/components/function-controller/migrator/pkg/resource/apis/types"
	"github.com/pkg/errors"
)

const (
	servicePort = 80
)

type ApiOperator struct {
	Data   *apiTypes.Api
	ResCli apis.Api
}

func (m migrator) createApirules() error {
	apiList, err := apis.New(m.dynamicCli, "", "", m.cfg.WaitTimeout, m.log.Info).List()
	if err != nil {
		return errors.Wrap(err, "while listing APIs")
	}

	for _, item := range apiList {
		cli := apirule.New(m.dynamicCli, item.Name, item.Namespace, m.cfg.WaitTimeout, m.log.Info)
		host := fmt.Sprintf("%s.%s", item.Name, m.cfg.Domain)

		m.log.WithValues("Name", item.Name,
			"Namespace", item.Namespace,
			"GroupVersion", apiruleTypes.GroupVersion.String(),
		).Info("Creating APIRule")

		if err := cli.Create(item.Name, host, servicePort); err != nil {
			return errors.Wrapf(err, "while creating ApiRule %s in namespace %s", item.Name, item.Namespace)
		}
	}

	return nil
}

func (m migrator) deleteApis() error {
	for _, fn := range m.kubelessFns {
		m.log.WithValues("name", fn.Data.Name,
			"namespace", fn.Data.Namespace,
			"GroupVersion", kubelessv1beta1.SchemeGroupVersion.String(),
		).Info("Deleting kubeless function")

		if err := fn.ResCli.Delete(); err != nil {
			return errors.Wrapf(err, "while deleting Kubeless function %s in namespace %s", fn.Data.Name, fn.Data.Namespace)
		}
	}
	return nil
}
