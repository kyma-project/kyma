package compass_runtime_agent

import (
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/applications"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/compassruntimeagentinit"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/director"
	"net/http"
	"testing"

	cli "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	"github.com/stretchr/testify/suite"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type CompassRuntimeAgentSuite struct {
	suite.Suite
	applicationsClientSet *cli.Clientset
	coreClientSet         *kubernetes.Clientset
	directorClient        director.Client
	appComparator         applications.Comparator
	cleanupFunction       compassruntimeagentinit.RollbackFunc
}

func initDirectorClient() director.Client {
	return nil
}

func (ts *CompassRuntimeAgentSuite) SetupSuite() {
	cfg, err := rest.InClusterConfig()
	ts.Require().Nil(err)

	ts.applicationsClientSet, err = cli.NewForConfig(cfg)
	ts.Require().Nil(err)

	ts.coreClientSet, err = kubernetes.NewForConfig(cfg)
	ts.Require().Nil(err)

	// TODO Init client
	ts.directorClient = director.NewDirectorClient(nil, nil)
	ts.Require().Nil(err)

	// TODO: Pass namespaces names
	secretComparator, err := applications.NewSecretComparator(ts.Require(), ts.coreClientSet, "", "")
	ts.Require().Nil(err)

	applicationGetter := ts.applicationsClientSet.ApplicationconnectorV1alpha1().Applications()
	ts.appComparator, err = applications.NewComparator(ts.Require(), secretComparator, applicationGetter, "", "")
	ts.Require().Nil(err)

	// TODO: Uncomment when directorClient satisfies the needed interface
	//ts.cleanupFunction, err = compassruntimeagentinit.NewCompassRuntimeAgentConfigurator(ts.directorClient, ts.coreClientSet, "tenant").Do("runtimeName")
	//ts.Require().Nil(err)
}

func (ts *CompassRuntimeAgentSuite) TearDownSuite() {
	if ts.cleanupFunction != nil {
		err := ts.cleanupFunction()
		ts.Nil(err)
	}

	_, err := http.Post("http://localhost:15000/quitquitquit", "", nil)
	ts.Nil(err)
	_, err = http.Post("http://localhost:15020/quitquitquit", "", nil)
	ts.Nil(err)
}

func TestCompassRuntimeAgentSuite(t *testing.T) {
	suite.Run(t, new(CompassRuntimeAgentSuite))
}
