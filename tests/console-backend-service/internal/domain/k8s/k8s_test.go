// +build acceptance

package k8s

import (
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service/pkg/exit"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"os"
	"testing"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/auth"
)

var AuthSuite *auth.TestSuite
const testNamespacePrefix string = "console-backend-service-k8s-"
var testNamespace string

func TestMain(m *testing.M) {
	k8sClient, _, err := client.NewClientWithConfig()
	exit.OnError(err, "while creating k8s client")

	namespace, err := k8sClient.Namespaces().Create(fixNamespace(testNamespacePrefix))
	exit.OnError(err, "while creating namespace")
	testNamespace = namespace.Name
	AuthSuite = auth.New()

	code := m.Run()

	err = k8sClient.Namespaces().Delete(testNamespace, &metav1.DeleteOptions{})
	if err != nil {
		log.Printf("Error while deleting %s namespace: %s", testNamespace, err.Error())
	}

	os.Exit(code)
}
