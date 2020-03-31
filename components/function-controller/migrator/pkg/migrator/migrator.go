package migrator

import (
	"time"

	"github.com/go-logr/logr"
	"github.com/kyma-project/kyma/components/function-controller/migrator/pkg/resource/apis"
	"github.com/kyma-project/kyma/components/function-controller/migrator/pkg/resource/kubeless"
	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Config struct {
	WaitTimeout time.Duration
	Domain      string
}

type migrator struct {
	dynamicCli  dynamic.Interface
	log         logr.Logger
	cfg         Config
	kubelessFns []FunctionOperator
	apis        []ApiOperator
}

func New(restConfig *rest.Config, cfg Config) (*migrator, error) {
	logf := log.Log.WithName("migrator")

	dynamicCli, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating K8s Dynamic client")
	}

	kubelessFnList, err := kubeless.New(dynamicCli, "", "", cfg.WaitTimeout, logf.Info).List()
	if err != nil {
		return nil, errors.Wrap(err, "while listing Kubeless functions")
	}

	var fnOperators []FunctionOperator
	for _, kubelessFn := range kubelessFnList {
		resCli := kubeless.New(dynamicCli, kubelessFn.Name, kubelessFn.Namespace, cfg.WaitTimeout, logf.Info)
		fnOperators = append(fnOperators, FunctionOperator{
			Data:   kubelessFn,
			ResCli: *resCli,
		})
	}

	apiList, err := apis.New(dynamicCli, "", "", cfg.WaitTimeout, logf.Info).List()
	if err != nil {
		return nil, errors.Wrap(err, "while listing apis.gateway.kyma-project.io")
	}

	var apiOperators []ApiOperator
	for _, item := range apiList {
		resCli := apis.New(dynamicCli, item.Name, item.Namespace, cfg.WaitTimeout, logf.Info)
		apiOperators = append(apiOperators, ApiOperator{
			Data:   item,
			ResCli: *resCli,
		})
	}

	return &migrator{
		dynamicCli:  dynamicCli,
		log:         logf,
		cfg:         cfg,
		kubelessFns: fnOperators,
		apis:        apiOperators,
	}, nil
}

func (m *migrator) Run() error {
	if err := m.updateServiceBindingUsages(); err != nil {
		return err
	}

	if err := m.updateTriggers(); err != nil {
		return err
	}

	if err := m.createApirules(); err != nil {
		return err
	}

	if err := m.deleteKubelessFunctions(); err != nil {
		return err
	}

	if err := m.createServerlessFns(); err != nil {
		return err
	}

	return nil
}
