package backend

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"testing"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/deployment"

	appsv1 "k8s.io/api/apps/v1"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	reconcilertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

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
)

const (
	useExistingCluster       = false
	attachControlPlaneOutput = false
	kymaSystemNamespace      = "kyma-system"
	timeout                  = 15 * time.Second
	pollingInterval          = 1 * time.Second
	eventingBackendName      = "eventing-backend"
	bebSecret1name           = "beb-secret-1"
	bebSecret2name           = "beb-secret-2"
)

var k8sClient client.Client
var testEnv *envtest.Environment

func TestGetSecretForPublisher(t *testing.T) {
	testCases := []struct {
		name           string
		messagingData  []byte
		namespaceData  []byte
		expectedSecret corev1.Secret
		expectedError  error
	}{
		{
			name: "with valid message and namepsace data",
			messagingData: []byte("[{		\"broker\": {			\"type\": \"sapmgw\"		},		\"oa2\": {			\"clientid\": \"clientid\",			\"clientsecret\": \"clientsecret\",			\"granttype\": \"client_credentials\",			\"tokenendpoint\": \"https://token\"		},		\"protocol\": [\"amqp10ws\"],		\"uri\": \"wss://amqp\"	}, {		\"broker\": {			\"type\": \"sapmgw\"		},		\"oa2\": {			\"clientid\": \"clientid\",			\"clientsecret\": \"clientsecret\",			\"granttype\": \"client_credentials\",			\"tokenendpoint\": \"https://token\"		},		\"protocol\": [\"amqp10ws\"],		\"uri\": \"wss://amqp\"	},	{		\"broker\": {			\"type\": \"saprestmgw\"		},		\"oa2\": {			\"clientid\": \"rest-clientid\",			\"clientsecret\": \"rest-client-secret\",			\"granttype\": \"client_credentials\",			\"tokenendpoint\": \"https://rest-token\"		},		\"protocol\": [\"httprest\"],		\"uri\": \"https://rest-messaging\"	}]"),
			namespaceData: []byte("valid/namespace"),
			expectedSecret: corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      deployment.PublisherName,
					Namespace: deployment.PublisherNamespace,
					Labels: map[string]string{
						deployment.AppLabelKey: deployment.PublisherName,
					},
				},
				StringData: map[string]string{
					"client-id":        "rest-clientid",
					"client-secret":    "rest-client-secret",
					"token-endpoint":   "https://rest-token?grant_type=client_credentials&response_type=token",
					"ems-publish-host": "https://rest-messaging",
					"ems-publish-url":  "https://rest-messaging/sap/ems/v1/events",
					"beb-namespace":    "valid/namespace",
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

func getSecret(message, namespace []byte) *corev1.Secret {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: deployment.PublisherName,
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

	// Use a "live" client to assert against the live state of the API server.
	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	syncPeriod := time.Second * 2
	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		SyncPeriod:         &syncPeriod,
		MetricsBindAddress: ":7071",
	})
	Expect(err).To(BeNil())

	natsCommander := TestNATSCommander{}
	bebCommander := TestBEBCommander{}

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

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

var _ = Describe("Backend Reconciliation Tests", func() {
	var testId = 0
	// enable me for debugging
	// SetDefaultEventuallyTimeout(time.Minute)
	// SetDefaultEventuallyPollingInterval(time.Second)

	AfterEach(func() {
		// increment the test id before each "It" block, which can be used to create unique objects
		// note: "AfterEach" is used in sync mode, so no need to use locks here
		testId++
	})

	When("Creating a Eventing Backend and no secret labeled for BEB is found", func() {
		It("Should start with NATS", func() {
			ctx := context.Background()
			eventingBackendName := "eventing-backend"
			backend := reconcilertesting.WithEventingBackend(eventingBackendName, kymaSystemNamespace)
			ensureEventingBackendCreated(ctx, backend)
			publisherProxy := new(appsv1.Deployment)
			// Expect
			getPublisherProxyDeployment(ctx, publisherProxy).Should(reconcilertesting.HaveStatusReady())
			getEventingBackend(ctx, backend).Should(reconcilertesting.HaveNATSBackendReady())
		})

	})

	When("A secret labeled for BEB is found", func() {
		It("Should switch from NATS to BEB", func() {
			ctx := context.Background()
			backend := reconcilertesting.WithEventingBackend(eventingBackendName, kymaSystemNamespace)
			publisherProxy := new(appsv1.Deployment)
			bebSecret := reconcilertesting.WithBEBMessagingSecret(bebSecret1name, kymaSystemNamespace)
			bebSecret.Labels = map[string]string{
				BEBBackendSecretLabelKey: BEBBackendSecretLabelValue,
			}
			ensureBEBSecretCreated(ctx, bebSecret)
			// Expect
			getPublisherProxySecret(ctx).Should(And(
				reconcilertesting.HaveValidClientID(deployment.PublisherSecretClientIDKey, "rest-clientid"),
				reconcilertesting.HaveValidClientSecret(deployment.PublisherSecretClientSecretKey, "rest-client-secret"),
				reconcilertesting.HaveValidTokenEndpoint(deployment.PublisherSecretTokenEndpointKey, "https://rest-token?grant_type=client_credentials&response_type=token"),
				reconcilertesting.HaveValidEMSPublishURL(PublisherSecretEMSHostKey, "https://rest-messaging"),
				reconcilertesting.HaveValidEMSPublishURL(deployment.PublisherSecretEMSURLKey, "https://rest-messaging/sap/ems/v1/events"),
				reconcilertesting.HaveValidBEBNamespace(deployment.PublisherSecretBEBNamespaceKey, "test/ns"),
			))
			getPublisherProxyDeployment(ctx, publisherProxy).Should(reconcilertesting.HaveStatusReady())
			getEventingBackend(ctx, backend).Should(reconcilertesting.HaveBEBBackendReady())
		})

	})

	When("More than one secret is found label for BEB usage", func() {
		It("Should take down eventing", func() {
			ctx := context.Background()
			backend := reconcilertesting.WithEventingBackend(eventingBackendName, kymaSystemNamespace)
			By("Creating second secret with the BEB label")
			bebSecret2 := reconcilertesting.WithBEBMessagingSecret(bebSecret2name, kymaSystemNamespace)
			bebSecret2.Labels = map[string]string{
				BEBBackendSecretLabelKey: BEBBackendSecretLabelValue,
			}
			ensureBEBSecretCreated(ctx, bebSecret2)

			By("Checking EventingReady status is set to false")
			updatedBackend := &eventingv1alpha1.EventingBackend{}
			Eventually(func() bool {
				lookupKey := types.NamespacedName{
					Namespace: backend.Namespace,
					Name:      backend.Name,
				}
				if err := k8sClient.Get(ctx, lookupKey, updatedBackend); err != nil {
					return false
				}
				if eventingReady := updatedBackend.Status.EventingReady; eventingReady == nil {
					return false
				} else {
					return *eventingReady
				}
			}, timeout, pollingInterval).Should(BeFalse())
		})
	})

	When("Two BEB secrets take down eventing and one secret is removed", func() {
		It("Should restore eventing status", func() {
			ctx := context.Background()
			backend := reconcilertesting.WithEventingBackend(eventingBackendName, kymaSystemNamespace)
			By("Deleting the second secret with the BEB label")
			bebSecret2 := reconcilertesting.WithBEBMessagingSecret(bebSecret2name, kymaSystemNamespace)
			Expect(k8sClient.Delete(ctx, bebSecret2)).Should(BeNil())

			By("Checking EventingReady status is set to true")
			updatedBackend := &eventingv1alpha1.EventingBackend{}
			Eventually(func() bool {
				lookupKey := types.NamespacedName{
					Namespace: backend.Namespace,
					Name:      backend.Name,
				}
				if err := k8sClient.Get(ctx, lookupKey, updatedBackend); err != nil {
					return false
				}
				if eventingReady := updatedBackend.Status.EventingReady; eventingReady == nil {
					return false
				} else {
					return *eventingReady
				}
			}, timeout, pollingInterval).Should(BeTrue())
			Expect(updatedBackend.Status.Backend).Should(Equal(eventingv1alpha1.BebBackendType))
		})
	})

	PWhen("Switching two NATS and then starting NATS controller fails", func() {
		It("Should mark Eventing Backend CR to not ready", func() {
			_ = context.Background()
			_ = fmt.Sprintf("sub-%d", testId)

		})
	})

	PWhen("Switching to BEB and then starting BEB controller fails", func() {
		It("Should mark Eventing Backend CR to not ready", func() {
			_ = context.Background()
			_ = fmt.Sprintf("sub-%d", testId)

		})
	})
})

func ensureEventingBackendCreated(ctx context.Context, backend *eventingv1alpha1.EventingBackend) {
	By(fmt.Sprintf("Ensuring the test namespace %q is created", kymaSystemNamespace))
	// create testing namespace
	namespace := reconcilertesting.WithNamespace(kymaSystemNamespace)
	err := k8sClient.Create(ctx, namespace)
	if !k8serrors.IsAlreadyExists(err) {
		Expect(err).ShouldNot(HaveOccurred())
	}

	By(fmt.Sprintf("Ensuring an Eventing Backend %q is created", backend.Name))
	err = k8sClient.Create(ctx, backend)
	if !k8serrors.IsAlreadyExists(err) {
		Expect(err).Should(BeNil())
	}
}

// getEventingBackend fetches an EventingBackend using the lookupKey and allows to make assertions on it
func getEventingBackend(ctx context.Context, evBackend *eventingv1alpha1.EventingBackend) AsyncAssertion {
	return Eventually(func() *eventingv1alpha1.EventingBackend {
		lookupKey := types.NamespacedName{
			Namespace: evBackend.Namespace,
			Name:      evBackend.Name,
		}
		if err := k8sClient.Get(ctx, lookupKey, evBackend); err != nil {
			log.Printf("failed to fetch eventingbackend(%s): %v", lookupKey.String(), err)
			return nil
		}
		log.Printf("[backend] name:%s ns:%s status:%v", evBackend.Name, evBackend.Namespace,
			evBackend.Status)
		return evBackend
	}, timeout, pollingInterval)
}

func ensureBEBSecretCreated(ctx context.Context, secret *corev1.Secret) {
	By(fmt.Sprintf("Ensuring the test namespace %q is created", secret.Namespace))
	// create testing namespace
	namespace := reconcilertesting.WithNamespace(secret.Namespace)
	err := k8sClient.Create(ctx, namespace)
	if !k8serrors.IsAlreadyExists(err) {
		Expect(err).ShouldNot(HaveOccurred())
	}

	By(fmt.Sprintf("Ensuring an BEB Secret %q is created", secret.Name))
	err = k8sClient.Create(ctx, secret)
	Expect(err).Should(BeNil())
}

// getPublisherProxyDeployment fetches PublisherProxy deployment
func getPublisherProxySecret(ctx context.Context) AsyncAssertion {
	return Eventually(func() *corev1.Secret {
		lookupKey := types.NamespacedName{
			Namespace: deployment.PublisherNamespace,
			Name:      deployment.PublisherName,
		}
		secret := new(corev1.Secret)
		if err := k8sClient.Get(ctx, lookupKey, secret); err != nil {
			log.Printf("failed to fetch publisher proxy secret(%s): %v", lookupKey.String(), err)
			return nil
		}
		return secret
	}, timeout, pollingInterval)
}

// getPublisherProxySecret fetches PublisherProxy deployment
func getPublisherProxyDeployment(ctx context.Context, publisher *appsv1.Deployment) AsyncAssertion {
	return Eventually(func() *appsv1.Deployment {
		lookupKey := types.NamespacedName{
			Namespace: deployment.PublisherNamespace,
			Name:      deployment.PublisherName,
		}
		if err := k8sClient.Get(ctx, lookupKey, publisher); err != nil {
			log.Printf("failed to fetch eventingbackend(%s): %v", lookupKey.String(), err)
			return nil
		}
		return publisher
	}, timeout, pollingInterval)
}

type TestNATSCommander struct{}

func (t TestNATSCommander) Init(mgr manager.Manager) error {
	return nil
}

func (t TestNATSCommander) Start() error {
	log.Printf("Test NATS Controller started!")
	return nil
}

func (t TestNATSCommander) Stop() error {
	return nil
}

type TestBEBCommander struct{}

func (t TestBEBCommander) Init(mgr manager.Manager) error {
	return nil
}

func (t TestBEBCommander) Start() error {
	log.Printf("Test BEB Controller started!")
	return nil
}

func (t TestBEBCommander) Stop() error {
	return nil
}
