package migrator

import (
	"time"

	"github.com/go-logr/logr"
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
}

func New(restConfig *rest.Config, cfg Config) (*migrator, error) {
	logf := log.Log.WithName("migrator")

	dynamicCli, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating K8s Dynamic client")
	}

	return &migrator{
		dynamicCli: dynamicCli,
		log:        logf,
		cfg:        cfg,
	}, nil
}

func (m *migrator) Run() error {
	kubelessFns, err := m.fetchKubelessData()
	if err != nil {
		return err
	}
	m.kubelessFns = kubelessFns

	if err := m.updateServiceBindingUsages(); err != nil {
		return err
	}

	if err := m.updateTriggers(); err != nil {
		return err
	}

	if err := m.createApirules(); err != nil {
		return err
	}
	// if err := m.deleteKubelessFunctions(); err != nil {
	// 	return err
	// }

	if err := m.createServerlessFns(); err != nil {
		return err
	}

	return nil
}
