package integration

import (
	"fmt"
	"github.com/kyma-project/kyma/components/console-backend-service/integration/graphql"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/app"
	"github.com/stretchr/testify/require"
	"github.com/vrischmann/envconfig"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"
	rbacClient "k8s.io/client-go/kubernetes/typed/rbac/v1"
	"k8s.io/client-go/rest"
	"net"
	"net/http"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sync"
	"testing"
)

var (
	restConfig    *rest.Config
	gqlEndpoint   string
	rbacClientset *rbacClient.RbacV1Client
)

func TestMain(m *testing.M) {
	// setup and start kube-apiserver
	var err error
	environment := &envtest.Environment{
		KubeAPIServerFlags: append([]string{"--authorization-mode=RBAC"}, envtest.DefaultKubeAPIServerFlags...),
	}
	restConfig, err = environment.Start()
	if err != nil {
		panic(err)
	}

	_, err = envtest.InstallCRDs(restConfig, envtest.CRDInstallOptions{
		Paths:              []string{"../../../resources/cluster-essentials/templates/crds"},
		ErrorIfPathMissing: true,
	})
	if err != nil {
		panic(err)
	}

	rbacClientset, err = rbacClient.NewForConfig(restConfig)
	if err != nil {
		panic(err)
	}

	appConfig := app.Config{}
	err = envconfig.InitWithOptions(&appConfig, envconfig.Options{Prefix: "CBS_TEST", AllOptional: true})
	if err != nil {
		panic(err)
	}

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	gqlEndpoint = fmt.Sprintf("http://127.0.0.1:%v/graphql", port)
	wg := &sync.WaitGroup{}
	stopCh := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		var auth authenticator.RequestFunc = AuthAlwaysSuccess
		err = app.Run(listener, stopCh, appConfig, restConfig, auth)
		if err != nil {
			panic(err)
		}
	}()

	status := m.Run()

	close(stopCh)
	wg.Wait()

	os.Exit(status)
}

func AuthAlwaysSuccess(req *http.Request) (*authenticator.Response, bool, error) {
	return &authenticator.Response{
		User: &user.DefaultInfo{
			Name: req.Header.Get("user"),
			UID:  "uid",
		},
	}, true, nil
}

func givenUserCanAccessResource(user, group, name string, verbs []string) error {
	roleName := fmt.Sprintf("%s-%s-%s", user, group, name)
	_, err := rbacClientset.ClusterRoles().Create(&rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: roleName,
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     verbs,
				APIGroups: []string{group},
				Resources: []string{name},
			},
		},
	})
	if err != nil {
		return err
	}

	_, err = rbacClientset.ClusterRoleBindings().Create(&rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: roleName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:     "User",
				APIGroup: "rbac.authorization.k8s.io",
				Name:     user,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     roleName,
		},
	})
	return err
}

func givenUserCannotAccessResource(user, group, name string) error {
	roleName := fmt.Sprintf("%s-%s-%s", user, group, name)
	err := rbacClientset.ClusterRoleBindings().Delete(roleName, &metav1.DeleteOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	err = rbacClientset.ClusterRoles().Delete(roleName, &metav1.DeleteOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}
	return nil
}

func thenRequestsShouldBeDenied(t *testing.T, gqlClient *graphql.Client, reqs ...*graphql.Request) {
	for _, req := range reqs {
		rsp := make(map[string]interface{})
		err := gqlClient.Do(req, &rsp)
		require.EqualError(t, err, "graphql: access denied")
	}
}
