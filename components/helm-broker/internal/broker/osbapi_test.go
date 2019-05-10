package broker_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/komkom/go-jsonhash"
	"github.com/pborman/uuid"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	rls "k8s.io/helm/pkg/proto/hapi/services"

	"github.com/kyma-project/kyma/components/helm-broker/platform/ptr"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/bind"
	"github.com/kyma-project/kyma/components/helm-broker/internal/broker"
	"github.com/kyma-project/kyma/components/helm-broker/internal/broker/automock"
	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle"
	bundle_automock "github.com/kyma-project/kyma/components/helm-broker/internal/bundle/automock"
	"github.com/kyma-project/kyma/components/helm-broker/internal/platform/logger/spy"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage"
)

func newOSBAPITestSuite(t *testing.T, docsTopicsService *bundle_automock.DocsTopicsService) *osbapiTestSuite {
	logSink := spy.NewLogSink()
	logSink.RawLogger.Out = ioutil.Discard

	sFact, err := storage.NewFactory(storage.NewConfigListAllMemory())
	require.NoError(t, err)

	ts := &osbapiTestSuite{
		t:              t,
		StorageFactory: sFact,
		HelmClient:     &automock.HelmClient{},
		LogSink:        logSink,
	}

	ts.Exp.Populate()

	ts.OperationIDProvider = func() (internal.OperationID, error) {
		return ts.Exp.OperationID, nil
	}

	ts.BrokerServer = broker.NewWithIDProvider(
		sFact.Bundle(),
		sFact.Chart(),
		sFact.InstanceOperation(),
		sFact.Instance(),
		sFact.InstanceBindData(),
		&fakeBindTmplRenderer{},
		&fakeBindTmplResolver{},
		ts.HelmClient,
		bundle.NewSyncer(sFact.Bundle(), sFact.Chart(), docsTopicsService, logSink.Logger),
		logSink.Logger, ts.OperationIDProvider)

	return ts
}

type osbapiTestSuite struct {
	t *testing.T

	BrokerServer        *broker.Server
	StorageFactory      storage.Factory
	HelmClient          *automock.HelmClient
	LogSink             *spy.LogSink
	OperationIDProvider func() (internal.OperationID, error)

	osbClient osb.Client

	serverWg     sync.WaitGroup
	serverCancel func()
	ServerAddr   string

	Exp expAll
}

func (ts *osbapiTestSuite) ServerRun() {
	ctx, cancel := context.WithCancel(context.Background())
	ts.serverWg.Add(1)

	startedCh := make(chan struct{})

	go func() {
		assert.Equal(ts.t, http.ErrServerClosed, ts.BrokerServer.Run(ctx, ":0", startedCh))
		ts.serverWg.Done()
	}()

	// TODO: wrap in timeout
	<-startedCh
	ts.ServerAddr = ts.BrokerServer.Addr()
	ts.serverCancel = cancel
}

func (ts *osbapiTestSuite) ServerShutdown() {
	ts.serverCancel()
	ts.serverWg.Wait()
}

func (ts *osbapiTestSuite) OSBClient() osb.Client {
	if ts.osbClient == nil {
		config := osb.DefaultClientConfiguration()
		config.URL = fmt.Sprintf("http://%s", ts.ServerAddr)

		osbClient, err := osb.NewClient(config)
		require.NoError(ts.t, err)
		ts.osbClient = osbClient
	}

	return ts.osbClient
}

func (ts *osbapiTestSuite) AssertOperationState(exp internal.OperationState) bool {

	doCheck := func() bool {
		op, err := ts.StorageFactory.InstanceOperation().Get(ts.Exp.InstanceID, ts.Exp.OperationID)
		require.NoError(ts.t, err)
		if op.State == exp {
			return true
		}
		return false
	}

	if doCheck() {
		return true
	}

	timeoutTotal := time.After(time.Second)
Polling:
	for {
		select {
		case <-timeoutTotal:
			ts.t.Error("timeout on instance operation state change")
			break Polling
		case <-time.After(time.Millisecond):
		}

		if doCheck() {
			return true
		}
	}

	return false
}

func TestOSBAPIStatusSuccess(t *testing.T) {
	// GIVEN
	ts := newOSBAPITestSuite(t, nil)
	ts.ServerRun()
	defer ts.ServerShutdown()

	client := &http.Client{}
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://%s/statusz", ts.ServerAddr), nil)
	req.Header.Set(osb.APIVersionHeader, "2.13")

	// WHEN
	resp, err := client.Do(req)

	// THEN
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestOSBAPICatalogSuccess(t *testing.T) {
	// GIVEN
	docsTopicsSvc := &bundle_automock.DocsTopicsService{}
	docsTopicsSvc.On("EnsureClusterDocsTopicRemoved", "fix-B-ID").Return(nil)
	defer docsTopicsSvc.AssertExpectations(t)

	ts := newOSBAPITestSuite(t, docsTopicsSvc)
	ts.ServerRun()
	defer ts.ServerShutdown()

	fixBundle := ts.Exp.NewBundle()
	ts.StorageFactory.Bundle().Upsert(fixBundle)

	// WHEN
	_, err := ts.OSBClient().GetCatalog()

	// THEN
	require.NoError(t, err)
}

func TestOSBAPIProvisionSuccess(t *testing.T) {
	// GIVEN
	ts := newOSBAPITestSuite(t, nil)

	ts.HelmClient.On("Install", mock.Anything, mock.Anything, ts.Exp.ReleaseName, ts.Exp.Namespace).Return(&rls.InstallReleaseResponse{}, nil).Once()
	defer ts.HelmClient.AssertExpectations(t)

	ts.ServerRun()
	defer ts.ServerShutdown()

	fixBundle := ts.Exp.NewBundle()
	ts.StorageFactory.Bundle().Upsert(fixBundle)

	fixChart := ts.Exp.NewChart()
	ts.StorageFactory.Chart().Upsert(fixChart)

	nsUID := uuid.NewRandom().String()
	req := &osb.ProvisionRequest{
		AcceptsIncomplete: true,
		InstanceID:        string(ts.Exp.InstanceID),
		ServiceID:         string(ts.Exp.Service.ID),
		PlanID:            string(ts.Exp.ServicePlan.ID),
		Context: map[string]interface{}{
			"namespace": string(ts.Exp.Namespace),
		},
		OrganizationGUID: nsUID,
		SpaceGUID:        nsUID,
	}

	// WHEN
	resp, err := ts.OSBClient().ProvisionInstance(req)

	// THEN
	require.NoError(t, err)

	require.True(t, resp.Async)
	assert.EqualValues(t, ts.Exp.OperationID, *resp.OperationKey)

	ts.AssertOperationState(internal.OperationStateSucceeded)
}

func TestOSBAPIProvisionRepeatedOnAlreadyFullyProvisionedInstance(t *testing.T) {
	// GIVEN
	ts := newOSBAPITestSuite(t, nil)

	fixInstance := ts.Exp.NewInstance()
	fixInstance.ParamsHash = jsonhash.HashS(map[string]interface{}{})
	ts.StorageFactory.Instance().Insert(fixInstance)

	fixOperation := ts.Exp.NewInstanceOperation(internal.OperationTypeCreate, internal.OperationStateSucceeded)
	ts.StorageFactory.InstanceOperation().Insert(fixOperation)

	ts.ServerRun()
	defer ts.ServerShutdown()

	nsUID := uuid.NewRandom().String()
	req := &osb.ProvisionRequest{
		AcceptsIncomplete: true,
		InstanceID:        string(ts.Exp.InstanceID),
		ServiceID:         string(ts.Exp.Service.ID),
		PlanID:            string(ts.Exp.ServicePlan.ID),
		Context: map[string]interface{}{
			"namespace": string(ts.Exp.Namespace),
		},
		OrganizationGUID: nsUID,
		SpaceGUID:        nsUID,
	}

	// WHEN
	resp, err := ts.OSBClient().ProvisionInstance(req)

	// THEN
	require.NoError(t, err)

	assert.False(t, resp.Async)
	assert.Nil(t, resp.OperationKey)

	// No activity on tiller should happen
	defer ts.HelmClient.AssertExpectations(t)
}

func TestOSBAPIProvisionRepeatedOnProvisioningInProgress(t *testing.T) {
	// GIVEN
	ts := newOSBAPITestSuite(t, nil)

	fixInstance := ts.Exp.NewInstance()
	fixInstance.ParamsHash = jsonhash.HashS(map[string]interface{}{})
	ts.StorageFactory.Instance().Insert(fixInstance)

	fixOperation := ts.Exp.NewInstanceOperation(internal.OperationTypeCreate, internal.OperationStateInProgress)
	expOpID := internal.OperationID("fix-op-id")
	fixOperation.OperationID = expOpID
	ts.StorageFactory.InstanceOperation().Insert(fixOperation)

	ts.ServerRun()
	defer ts.ServerShutdown()

	nsUID := uuid.NewRandom().String()
	req := &osb.ProvisionRequest{
		AcceptsIncomplete: true,
		InstanceID:        string(ts.Exp.InstanceID),
		ServiceID:         string(ts.Exp.Service.ID),
		PlanID:            string(ts.Exp.ServicePlan.ID),
		Context: map[string]interface{}{
			"namespace": string(ts.Exp.Namespace),
		},
		OrganizationGUID: nsUID,
		SpaceGUID:        nsUID,
	}

	// WHEN
	resp, err := ts.OSBClient().ProvisionInstance(req)

	// THEN
	require.NoError(t, err)

	assert.True(t, resp.Async)
	assert.EqualValues(t, expOpID, *resp.OperationKey)

	// No activity on tiller should happen
	defer ts.HelmClient.AssertExpectations(t)
}

func TestOSBAPIProvisionConflictErrorOnAlreadyFullyProvisionedInstance(t *testing.T) {
	// GIVEN
	ts := newOSBAPITestSuite(t, nil)

	fixInstance := ts.Exp.NewInstance()
	ts.StorageFactory.Instance().Insert(fixInstance)

	fixOperation := ts.Exp.NewInstanceOperation(internal.OperationTypeCreate, internal.OperationStateSucceeded)
	ts.StorageFactory.InstanceOperation().Insert(fixOperation)

	ts.ServerRun()
	defer ts.ServerShutdown()

	nsUID := uuid.NewRandom().String()
	req := &osb.ProvisionRequest{
		AcceptsIncomplete: true,
		InstanceID:        string(ts.Exp.InstanceID),
		ServiceID:         string(ts.Exp.Service.ID),
		PlanID:            string(ts.Exp.ServicePlan.ID),
		Context: map[string]interface{}{
			"namespace": string(ts.Exp.Namespace),
		},
		OrganizationGUID: nsUID,
		SpaceGUID:        nsUID,
	}

	// WHEN
	resp, err := ts.OSBClient().ProvisionInstance(req)

	// THEN
	assert.Nil(t, resp)
	assert.Equal(t, osb.HTTPStatusCodeError{StatusCode:http.StatusConflict, ErrorMessage:ptrStr(fmt.Sprintf("while comparing provisioning parameters map[]: provisioning parameters hash differs - new hqB_njX1ZeC2Y0XVG9_uBw==, old TODO, for instance fix-I-ID")), Description:ptrStr("")}, err)

	// No activity on tiller should happen
	defer ts.HelmClient.AssertExpectations(t)
}

func TestOSBAPIProvisionConflictErrorOnProvisioningInProgress(t *testing.T) {
	// GIVEN
	ts := newOSBAPITestSuite(t, nil)

	fixInstance := ts.Exp.NewInstance()
	ts.StorageFactory.Instance().Insert(fixInstance)

	fixOperation := ts.Exp.NewInstanceOperation(internal.OperationTypeCreate, internal.OperationStateInProgress)
	expOpID := internal.OperationID("fix-op-id")
	fixOperation.OperationID = expOpID
	ts.StorageFactory.InstanceOperation().Insert(fixOperation)

	ts.ServerRun()
	defer ts.ServerShutdown()

	nsUID := uuid.NewRandom().String()
	req := &osb.ProvisionRequest{
		AcceptsIncomplete: true,
		InstanceID:        string(ts.Exp.InstanceID),
		ServiceID:         string(ts.Exp.Service.ID),
		PlanID:            string(ts.Exp.ServicePlan.ID),
		Context: map[string]interface{}{
			"namespace": string(ts.Exp.Namespace),
		},
		OrganizationGUID: nsUID,
		SpaceGUID:        nsUID,
	}

	// WHEN
	resp, err := ts.OSBClient().ProvisionInstance(req)

	// THEN
	assert.Nil(t, resp)
	assert.Equal(t, osb.HTTPStatusCodeError{StatusCode:http.StatusConflict, ErrorMessage:ptrStr(fmt.Sprintf("while comparing provisioning parameters map[]: provisioning parameters hash differs - new hqB_njX1ZeC2Y0XVG9_uBw==, old TODO, for instance fix-I-ID")), Description:ptrStr("")}, err)

	// No activity on tiller should happen
	defer ts.HelmClient.AssertExpectations(t)
}

func TestOSBAPIDeprovisionOnAlreadyDeprovisionedInstance(t *testing.T) {
	// GIVEN
	ts := newOSBAPITestSuite(t, nil)

	fixInstance := ts.Exp.NewInstance()
	ts.StorageFactory.Instance().Insert(fixInstance)

	fixOperation := ts.Exp.NewInstanceOperation(internal.OperationTypeRemove, internal.OperationStateSucceeded)
	ts.StorageFactory.InstanceOperation().Insert(fixOperation)

	ts.ServerRun()
	defer ts.ServerShutdown()

	req := &osb.DeprovisionRequest{
		AcceptsIncomplete: true,
		InstanceID:        string(ts.Exp.InstanceID),
		ServiceID:         string(ts.Exp.Service.ID),
		PlanID:            string(ts.Exp.ServicePlan.ID),
	}

	// WHEN
	resp, err := ts.OSBClient().DeprovisionInstance(req)

	// THEN
	require.NoError(t, err)

	assert.False(t, resp.Async)
	assert.Nil(t, resp.OperationKey)

	// No activity on tiller should happen
	defer ts.HelmClient.AssertExpectations(t)
}

func TestOSBAPIDeprovisionOnAlreadyDeprovisionedAndRemovedInstance(t *testing.T) {
	// GIVEN
	ts := newOSBAPITestSuite(t, nil)
	// storage does not contain any data

	ts.ServerRun()
	defer ts.ServerShutdown()

	req := &osb.DeprovisionRequest{
		AcceptsIncomplete: true,
		InstanceID:        string(ts.Exp.InstanceID),
		ServiceID:         string(ts.Exp.Service.ID),
		PlanID:            string(ts.Exp.ServicePlan.ID),
	}

	// WHEN
	resp, err := ts.OSBClient().DeprovisionInstance(req)

	// THEN
	require.NoError(t, err)

	assert.False(t, resp.Async)
	assert.Nil(t, resp.OperationKey)

	// No activity on tiller should happen
	defer ts.HelmClient.AssertExpectations(t)
}

func TestOSBAPIDeprovisionRepeatedOnDeprovisioningInProgress(t *testing.T) {
	// GIVEN
	ts := newOSBAPITestSuite(t, nil)

	fixInstance := ts.Exp.NewInstance()
	ts.StorageFactory.Instance().Insert(fixInstance)

	fixOperation := ts.Exp.NewInstanceOperation(internal.OperationTypeRemove, internal.OperationStateInProgress)
	expOpID := internal.OperationID("fix-op-id")
	fixOperation.OperationID = expOpID
	ts.StorageFactory.InstanceOperation().Insert(fixOperation)

	ts.ServerRun()
	defer ts.ServerShutdown()

	req := &osb.DeprovisionRequest{
		AcceptsIncomplete: true,
		InstanceID:        string(ts.Exp.InstanceID),
		ServiceID:         string(ts.Exp.Service.ID),
		PlanID:            string(ts.Exp.ServicePlan.ID),
	}

	// WHEN
	resp, err := ts.OSBClient().DeprovisionInstance(req)

	// THEN
	require.NoError(t, err)

	assert.True(t, resp.Async)
	assert.EqualValues(t, expOpID, *resp.OperationKey)

	// No activity on tiller should happen
	defer ts.HelmClient.AssertExpectations(t)
}

func TestOSBAPIDeprovisionSuccess(t *testing.T) {
	// GIVEN
	ts := newOSBAPITestSuite(t, nil)

	fixOperation := ts.Exp.NewInstanceOperation(internal.OperationTypeCreate, internal.OperationStateSucceeded)
	expOpID := internal.OperationID("fix-op-id")
	fixOperation.OperationID = expOpID
	ts.StorageFactory.InstanceOperation().Insert(fixOperation)

	ts.HelmClient.On("Delete", ts.Exp.ReleaseName).Return(nil).Once()
	defer ts.HelmClient.AssertExpectations(t)

	ts.ServerRun()
	defer ts.ServerShutdown()

	fixInstance := ts.Exp.NewInstance()
	ts.StorageFactory.Instance().Insert(fixInstance)

	req := &osb.DeprovisionRequest{
		AcceptsIncomplete: true,
		InstanceID:        string(ts.Exp.InstanceID),
		ServiceID:         string(ts.Exp.Service.ID),
		PlanID:            string(ts.Exp.ServicePlan.ID),
	}

	// WHEN
	resp, err := ts.OSBClient().DeprovisionInstance(req)

	// THEN
	require.NoError(t, err)

	require.True(t, resp.Async)
	assert.EqualValues(t, ts.Exp.OperationID, *resp.OperationKey)

	ts.AssertOperationState(internal.OperationStateSucceeded)
}

func TestOSBAPILastOperationSuccess(t *testing.T) {
	// GIVEN
	ts := newOSBAPITestSuite(t, nil)
	ts.ServerRun()
	defer ts.ServerShutdown()

	fixBundle := ts.Exp.NewBundle()
	ts.StorageFactory.Bundle().Upsert(fixBundle)

	fixInstance := ts.Exp.NewInstance()
	ts.StorageFactory.Instance().Insert(fixInstance)

	fixOperation := ts.Exp.NewInstanceOperation(internal.OperationTypeCreate, internal.OperationStateInProgress)
	ts.StorageFactory.InstanceOperation().Insert(fixOperation)

	// WHEN
	opKey := osb.OperationKey(ts.Exp.OperationID)
	req := &osb.LastOperationRequest{
		InstanceID:   string(ts.Exp.InstanceID),
		ServiceID:    ptr.String(string(ts.Exp.Service.ID)),
		PlanID:       ptr.String(string(ts.Exp.ServicePlan.ID)),
		OperationKey: &opKey,
	}
	resp, err := ts.OSBClient().PollLastOperation(req)

	// THEN
	require.NoError(t, err)
	assert.EqualValues(t, internal.OperationStateInProgress, resp.State)
	// TODO: match desc
}

func TestOSBAPILastOperationForNonExistingInstance(t *testing.T) {
	// GIVEN
	ts := newOSBAPITestSuite(t, nil)
	ts.ServerRun()
	defer ts.ServerShutdown()

	fixBundle := ts.Exp.NewBundle()
	ts.StorageFactory.Bundle().Upsert(fixBundle)

	// WHEN
	opKey := osb.OperationKey(ts.Exp.OperationID)
	req := &osb.LastOperationRequest{
		InstanceID:   string(ts.Exp.InstanceID),
		ServiceID:    ptr.String(string(ts.Exp.Service.ID)),
		PlanID:       ptr.String(string(ts.Exp.ServicePlan.ID)),
		OperationKey: &opKey,
	}
	_, err := ts.OSBClient().PollLastOperation(req)

	// THEN
	assert.True(t, osb.IsGoneError(err))
}

func TestOSBAPIBindFailureWithDisallowedParametersFieldInReq(t *testing.T) {
	// GIVEN
	ts := newOSBAPITestSuite(t, nil)
	ts.ServerRun()
	defer ts.ServerShutdown()

	fixBundle := ts.Exp.NewBundle()
	ts.StorageFactory.Bundle().Upsert(fixBundle)

	// WHEN
	req := &osb.BindRequest{
		BindingID:  "bind-id",
		InstanceID: "instance-id",
		ServiceID:  "svc-id",
		PlanID:     "bind-id",
		Parameters: map[string]interface{}{
			"params": "set-but-not-allowed",
		},
	}
	_, err := ts.OSBClient().Bind(req)

	// THEN
	require.Error(t, err)
	castedErr, ok := osb.IsHTTPError(err)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, castedErr.StatusCode)
}

type fakeBindTmplRenderer struct{}

func (fakeBindTmplRenderer) Render(bindTemplate internal.BundlePlanBindTemplate, resp *rls.InstallReleaseResponse) (bind.RenderedBindYAML, error) {
	return []byte(`fake`), nil
}

type fakeBindTmplResolver struct{}

func (fakeBindTmplResolver) Resolve(bindYAML bind.RenderedBindYAML, ns internal.Namespace) (*bind.ResolveOutput, error) {
	return &bind.ResolveOutput{}, nil
}

func ptrStr(str string) *string {
	return &str
}