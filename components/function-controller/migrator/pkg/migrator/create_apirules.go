package migrator

import (
	"fmt"

	"github.com/kyma-project/kyma/components/function-controller/migrator/pkg/resource/apirule"
	apiruleTypes "github.com/kyma-project/kyma/components/function-controller/migrator/pkg/resource/apirule/types"
	"github.com/pkg/errors"
)

const (
	servicePort = 80
)

func (m migrator) createApirules() error {
	for _, fn := range m.kubelessFns {
		cli := apirule.New(m.dynamicCli, fn.Data.Name, fn.Data.Namespace, m.cfg.WaitTimeout, m.log.Info)
		host := fmt.Sprintf("%s.%s", fn.Data.Name, m.cfg.Domain)

		m.log.WithValues("Name", fn.Data.Name,
			"Namespace", fn.Data.Namespace,
			"GroupVersion", apiruleTypes.GroupVersion.String(),
		).Info("Creating APIRule")

		if err := cli.Create(fn.Data.Name, host, servicePort); err != nil {
			return errors.Wrapf(err, "while creating ApiRule %s in namespace %s", fn.Data.Name, fn.Data.Namespace)
		}
	}

	return nil
}
