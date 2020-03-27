package migrator

import (
	"github.com/kyma-project/kyma/components/function-controller/migrator/pkg/resource/servicecatalog/servicebindingusage"
	"github.com/kyma-project/kyma/components/function-controller/migrator/pkg/resource/servicecatalog/types/v1alpha1"
	"github.com/pkg/errors"
)

const (
	knativeServiceKind   = "knative-service"
	kubelessFunctionKind = "function"
)

func (m migrator) updateServiceBindingUsages() error {
	sbus, err := m.getServiceBindingUsageList()
	if err != nil {
		return err
	}

	for _, sbu := range sbus {
		m.log.WithValues("Name", sbu.Data.Name,
			"Namespace", sbu.Data.Namespace,
			"GroupVersion", v1alpha1.SchemeGroupVersion.String(),
		).Info("Updating ServiceBindingUsage")

		sbu.Data.Spec.UsedBy.Kind = knativeServiceKind

		if err := sbu.ResCli.Update(sbu.Data); err != nil {
			return err
		}
	}
	return nil
}

type ServiceBindingUsageOperator struct {
	Data   *v1alpha1.ServiceBindingUsage
	ResCli *servicebindingusage.ServiceBindingUsage
}

func (m migrator) getServiceBindingUsageList() ([]ServiceBindingUsageOperator, error) {
	sbuList, err := servicebindingusage.New(m.dynamicCli, "", "", m.cfg.WaitTimeout, m.log.Info).List()
	if err != nil {
		return nil, errors.Wrap(err, "while listing ServiceBindingUsages")
	}

	var ret []ServiceBindingUsageOperator
	for _, sbu := range sbuList {
		if sbu.Spec.UsedBy.Kind != kubelessFunctionKind {
			continue
		}

		ret = append(ret, ServiceBindingUsageOperator{
			Data:   sbu,
			ResCli: servicebindingusage.New(m.dynamicCli, sbu.Name, sbu.Namespace, m.cfg.WaitTimeout, m.log.Info),
		})
	}

	return ret, nil
}
