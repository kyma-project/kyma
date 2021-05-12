package backend

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/commander"

	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/stretchr/testify/assert"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/api/core/v1"
)

const (
	useExistingCluster       = false
	attachControlPlaneOutput = false
)

var k8sClient client.Client
var testEnv *envtest.Environment

func TestGetSecretForPublisher(t *testing.T) {
	testCases := []struct {
		name           string
		messagingData  []byte
		namespaceData  []byte
		expectedSecret v1.Secret
		expectedError  error
	}{
		{
			name: "with valid message and namepsace data",
			messagingData: []byte("[{		\"broker\": {			\"type\": \"sapmgw\"		},		\"oa2\": {			\"clientid\": \"clientid\",			\"clientsecret\": \"clientsecret\",			\"granttype\": \"client_credentials\",			\"tokenendpoint\": \"https://token\"		},		\"protocol\": [\"amqp10ws\"],		\"uri\": \"wss://amqp\"	}, {		\"broker\": {			\"type\": \"sapmgw\"		},		\"oa2\": {			\"clientid\": \"clientid\",			\"clientsecret\": \"clientsecret\",			\"granttype\": \"client_credentials\",			\"tokenendpoint\": \"https://token\"		},		\"protocol\": [\"amqp10ws\"],		\"uri\": \"wss://amqp\"	},	{		\"broker\": {			\"type\": \"saprestmgw\"		},		\"oa2\": {			\"clientid\": \"rest-clientid\",			\"clientsecret\": \"rest-client-secret\",			\"granttype\": \"client_credentials\",			\"tokenendpoint\": \"https://rest-token\"		},		\"protocol\": [\"httprest\"],		\"uri\": \"https://rest-messaging\"	}]"),
			namespaceData: []byte("valid/namespace"),
			expectedSecret: v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      PublisherName,
					Namespace: PublisherNamespace,
					Labels: map[string]string{
						AppLabelKey: PublisherName,
					},
				},
				StringData: map[string]string{
					"client-id":       "rest-clientid",
					"client-secret":   "rest-client-secret",
					"token-endpoint":  "https://rest-token?grant_type=client_credentials&response_type=token",
					"ems-publish-url": "https://rest-messaging",
					"beb-namespace":   "valid/namespace",
				},
			},
		},
		{
			name:          "with empty message data",
			namespaceData: []byte("valid/namespace"),
			expectedError: errors.New("message is missing from BEB secret"),
		},
		{
			name: "with empty namespace data",
			messagingData: []byte("[{		\"broker\": {			\"type\": \"sapmgw\"		},		\"oa2\": {			\"clientid\": \"clientid\",			\"clientsecret\": \"clientsecret\",			\"granttype\": \"client_credentials\",			\"tokenendpoint\": \"https://token\"		},		\"protocol\": [\"amqp10ws\"],		\"uri\": \"wss://amqp\"	}, {		\"broker\": {			\"type\": \"sapmgw\"		},		\"oa2\": {			\"clientid\": \"clientid\",			\"clientsecret\": \"clientsecret\",			\"granttype\": \"client_credentials\",			\"tokenendpoint\": \"https://token\"		},		\"protocol\": [\"amqp10ws\"],		\"uri\": \"wss://amqp\"	},	{		\"broker\": {			\"type\": \"saprestmgw\"		},		\"oa2\": {			\"clientid\": \"rest-clientid\",			\"clientsecret\": \"rest-client-secret\",			\"granttype\": \"client_credentials\",			\"tokenendpoint\": \"https://rest-token\"		},		\"protocol\": [\"httprest\"],		\"uri\": \"https://rest-messaging\"	}]"),
			expectedError: errors.New("namespace is missing from BEB secret"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			publisherSecret := getSecret(tc.messagingData, tc.namespaceData)

			gotPublisherSecret, err := getSecretForPublisher(publisherSecret)
			if tc.expectedError != nil {
				assert.Equal(t, tc.expectedError.Error(), err.Error(), "invalid error")
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedSecret, *gotPublisherSecret, "invalid publisher secret")
		})
	}
}

func getSecret(message, namespace []byte) *v1.Secret {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: PublisherName,
		},
	}

	secret.Data = make(map[string][]byte)

	if len(message) > 0 {
		secret.Data["messaging"] = message
	}
	if len(namespace) > 0 {
		secret.Data["namespace"] = namespace
	}

	return secret
}

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "Eventing Backend Controller Suite", []Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter)))

	By("bootstrapping test environment")
	useExistingCluster := useExistingCluster

	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("../../", "config", "crd", "bases"),
			filepath.Join("../../", "config", "crd", "external"),
		},
		AttachControlPlaneOutput: attachControlPlaneOutput,
		UseExistingCluster:       &useExistingCluster,
	}

	var err error

	cfg, err := testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = eventingv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	// +kubebuilder:scaffold:scheme

	syncPeriod := time.Second * 2
	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		SyncPeriod:         &syncPeriod,
		MetricsBindAddress: ":7071",
	})
	Expect(err).ToNot(HaveOccurred())
	var natsCommander, bebCommander commander.Commander

	// Set env vars
	err = NewReconciler(
		context.Background(),
		natsCommander,
		bebCommander,
		k8sManager.GetClient(),
		k8sManager.GetCache(),
		ctrl.Log.WithName("reconciler").WithName("eventing-backend"),
		k8sManager.GetEventRecorderFor("backend-controller"),
	).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

	k8sClient = k8sManager.GetClient()
	Expect(k8sClient).ToNot(BeNil())

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

var _ = Describe("NATS Subscription Reconciliation Tests", func() {
	var testId = 0
	//var namespaceName = "test"

	// enable me for debugging
	// SetDefaultEventuallyTimeout(time.Minute)
	// SetDefaultEventuallyPollingInterval(time.Second)

	AfterEach(func() {
		// increment the test id before each "It" block, which can be used to create unique objects
		// note: "AfterEach" is used in sync mode, so no need to use locks here
		testId++

		// print eventing backend CR

		// print publisher proxy deployment

	})

	When("Creating a Eventing Backend without a secret", func() {
		It("Should create eventing infra for NATS", func() {
			_ = context.Background()
			_ = fmt.Sprintf("sub-%d", testId)

		})
	})

	When("Creating a secret with BEB label", func() {
		It("Should create eventing infra for BEB and create/update Eventing Backend", func() {
			_ = context.Background()
			_ = fmt.Sprintf("sub-%d", testId)

		})
	})

	When("Creating a eventing backend CR and then labelling a secret with BEB", func() {
		It("Should change eventing infra from NATS and BEB", func() {
			_ = context.Background()
			_ = fmt.Sprintf("sub-%d", testId)

		})
	})

	When("Creating a eventing backend CR and then labelling 2 secrets with BEB", func() {
		It("Should mark Eventing Backend CR to not ready", func() {
			_ = context.Background()
			_ = fmt.Sprintf("sub-%d", testId)

		})
	})

})
