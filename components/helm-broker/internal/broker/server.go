package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	negronilogrus "github.com/meatballhat/negroni-logrus"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"

	osb "github.com/pmorie/go-open-service-broker-client/v2"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

//go:generate mockery -name=catalogGetter -output=automock -outpkg=automock -case=underscore

type (
	catalogGetter interface {
		GetCatalog(ctx context.Context, osbCtx OsbContext) (*osb.CatalogResponse, error)
	}

	provisioner interface {
		Provision(ctx context.Context, osbCtx OsbContext, req *osb.ProvisionRequest) (*osb.ProvisionResponse, error)
	}

	deprovisioner interface {
		Deprovision(ctx context.Context, osbCtx OsbContext, req *osb.DeprovisionRequest) (*osb.DeprovisionResponse, error)
	}

	binder interface {
		Bind(ctx context.Context, osbCtx OsbContext, req *osb.BindRequest) (*osb.BindResponse, error)
	}

	unbinder interface {
		Unbind(ctx context.Context, osbCtx OsbContext, req *osb.UnbindRequest) (*osb.UnbindResponse, error)
	}

	lastOpGetter interface {
		GetLastOperation(ctx context.Context, osbCtx OsbContext, req *osb.LastOperationRequest) (*osb.LastOperationResponse, error)
	}
)

// Server implements HTTP server used to serve OSB API for helm broker.
type Server struct {
	catalogGetter catalogGetter
	provisioner   provisioner
	deprovisioner deprovisioner
	binder        binder
	unbinder      unbinder
	lastOpGetter  lastOpGetter
	logger        *logrus.Entry
	addr          string
}

// Addr returns address server is listening on.
// Its use is targeted for cases when address is not known, e.g. tests.
func (srv *Server) Addr() string {
	if srv.addr == "" {
		timer := time.NewTicker(time.Millisecond)
	waitLoop:
		for {
			<-timer.C

			if srv.addr != "" {
				break waitLoop
			}
		}
	}

	return srv.addr
}

// Run is starting HTTP server
func (srv *Server) Run(ctx context.Context, addr string, startedCh chan struct{}) error {
	listenAndServe := func(httpSrv *http.Server) error {
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return err
		}
		close(startedCh)
		lnTCP := ln.(*net.TCPListener)

		srv.addr = lnTCP.Addr().String()

		// TODO: add support for tcpKeepAliveListener
		return httpSrv.Serve(ln)
	}

	return srv.run(ctx, addr, listenAndServe)
}

// RunTLS is starting TLS server
func RunTLS(ctx context.Context, addr string, cert string, key string) error {
	return errors.New("TLS is not yet implemented")
}

// TODO: rewrite to go-sdk implementation with app and services
func (srv *Server) run(ctx context.Context, addr string, listenAndServe func(srv *http.Server) error) error {
	httpSrv := &http.Server{
		Addr:    addr,
		Handler: srv.createHandler(),
	}
	go func() {
		<-ctx.Done()
		c, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if httpSrv.Shutdown(c) != nil {
			httpSrv.Close()
		}
	}()
	return listenAndServe(httpSrv)
}

func (srv *Server) createHandler() http.Handler {
	var rtr = mux.NewRouter()

	// TODO: middleware: validate osbCtx.APIVersion that matches 2.12
	// TODO: middleware: add support for osbCtx.OriginatingIdentity

	rtr.HandleFunc("/statusz", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	}).Methods("GET")

	// sync operations
	rtr.HandleFunc("/v2/catalog", srv.catalogAction).Methods("GET")
	rtr.HandleFunc("/v2/service_instances/{instance_id}/last_operation", srv.getServiceInstanceLastOperationAction).Methods("GET")
	rtr.HandleFunc("/v2/service_instances/{instance_id}/service_bindings/{binding_id}", srv.bindAction).Methods("PUT")
	rtr.HandleFunc("/v2/service_instances/{instance_id}/service_bindings/{binding_id}", srv.unBindAction).Methods("DELETE")

	// async operations
	rtr.Path("/v2/service_instances/{instance_id}").Methods(http.MethodPut).Handler(
		negroni.New(&RequireAsyncMiddleware{}, negroni.WrapFunc(srv.provisionAction)),
	)
	rtr.Path("/v2/service_instances/{instance_id}").Methods(http.MethodDelete).Handler(
		negroni.New(&RequireAsyncMiddleware{}, negroni.WrapFunc(srv.deprovisionAction)),
	)

	logMiddleware := negronilogrus.NewMiddlewareFromLogger(srv.logger.Logger, "")
	logMiddleware.After = func(in *logrus.Entry, rw negroni.ResponseWriter, latency time.Duration, s string) *logrus.Entry {
		return in.WithFields(logrus.Fields{
			"status": rw.Status(),
			"took":   latency,
			"size":   rw.Size(),
		})
	}

	n := negroni.New(negroni.NewRecovery(), logMiddleware)
	n.Use(&OSBContextMiddleware{})
	n.UseHandler(rtr)
	return n
}

func (srv *Server) catalogAction(w http.ResponseWriter, r *http.Request) {
	osbCtx, _ := osbContextFromContext(r.Context())
	resp, err := srv.catalogGetter.GetCatalog(r.Context(), osbCtx)
	if err != nil {
		srv.writeErrorResponse(w, http.StatusBadRequest, err.Error(), "")
		return
	}

	if srv.logger != nil {
		srv.logger.WithFields(logrus.Fields{
			"action":              "catalog",
			"resp:services:count": len(resp.Services),
		}).Info("action response")
	}

	srv.writeResponse(w, http.StatusOK, resp)
}

func (srv *Server) provisionAction(w http.ResponseWriter, r *http.Request) {
	osbCtx, _ := osbContextFromContext(r.Context())

	var inDTO ProvisionRequestDTO

	if err := httpBodyToDTO(r, &inDTO); err != nil {
		srv.writeErrorResponse(w, http.StatusBadRequest, err.Error(), "")
	}

	instanceID := mux.Vars(r)["instance_id"]

	sReq := osb.ProvisionRequest{
		AcceptsIncomplete: true,
		InstanceID:        string(instanceID),
		ServiceID:         string(inDTO.ServiceID),
		PlanID:            string(inDTO.PlanID),
		OrganizationGUID:  inDTO.OrganizationGUID,
		SpaceGUID:         inDTO.SpaceGUID,
		Parameters:        inDTO.Parameters,
		Context: map[string]interface{}{
			"namespace": string(inDTO.Context.Namespace),
		},
	}

	sResp, err := srv.provisioner.Provision(r.Context(), osbCtx, &sReq)
	if err != nil {
		srv.writeErrorResponse(w, http.StatusBadRequest, err.Error(), "")
		return
	}

	logRespFields := logrus.Fields{
		"action":     "provision",
		"resp:async": sResp.Async,
	}
	logResp := func(fields logrus.Fields) {
		if srv.logger != nil {
			srv.logger.WithFields(fields).Info("action response")
		}
	}

	if !sResp.Async {
		logResp(logRespFields)
		srv.writeResponse(w, http.StatusOK, map[string]interface{}{})
		return
	}

	opID := internal.OperationID(*sResp.OperationKey)
	egDTO := ProvisionSuccessResponseDTO{
		Operation: &opID,
	}

	logRespFields["resp:operation:id"] = opID
	logResp(logRespFields)

	srv.writeResponse(w, http.StatusAccepted, egDTO)
}

func (srv *Server) deprovisionAction(w http.ResponseWriter, r *http.Request) {
	osbCtx, _ := osbContextFromContext(r.Context())

	instanceID := mux.Vars(r)["instance_id"]

	q := r.URL.Query()

	svcIDRaw := q.Get("service_id")
	planIDRaw := q.Get("plan_id")
	sReq := osb.DeprovisionRequest{
		AcceptsIncomplete: true,
		InstanceID:        string(instanceID),
		ServiceID:         svcIDRaw,
		PlanID:            planIDRaw,
	}

	sResp, err := srv.deprovisioner.Deprovision(r.Context(), osbCtx, &sReq)
	switch {
	case IsNotFoundError(err):
		srv.writeResponse(w, http.StatusGone, map[string]interface{}{})
		return
	case err != nil:
		srv.writeErrorResponse(w, http.StatusBadRequest, err.Error(), "")
		return
	}

	logRespFields := logrus.Fields{
		"action":     "deprovision",
		"resp:async": sResp.Async,
	}
	logResp := func(fields logrus.Fields) {
		if srv.logger != nil {
			srv.logger.WithFields(fields).Info("action response")
		}
	}

	if !sResp.Async {
		logResp(logRespFields)
		srv.writeResponse(w, http.StatusGone, map[string]interface{}{})
		return
	}

	opID := internal.OperationID(*sResp.OperationKey)
	egDTO := ProvisionSuccessResponseDTO{
		Operation: &opID,
	}

	logRespFields["resp:operation:id"] = opID
	logResp(logRespFields)

	srv.writeResponse(w, http.StatusAccepted, egDTO)
}

func (srv *Server) getServiceInstanceLastOperationAction(w http.ResponseWriter, r *http.Request) {
	osbCtx, _ := osbContextFromContext(r.Context())

	instanceID := mux.Vars(r)["instance_id"]
	var operationID internal.OperationID

	q := r.URL.Query()

	sReq := osb.LastOperationRequest{
		InstanceID: string(instanceID),
	}
	if svcIDRaw := q.Get("service_id"); svcIDRaw != "" {
		svcID := svcIDRaw
		sReq.ServiceID = &svcID
	}
	if planIDRaw := q.Get("plan_id"); planIDRaw != "" {
		planID := planIDRaw
		sReq.PlanID = &planID
	}
	if opIDRaw := q.Get("operation"); opIDRaw != "" {
		operationID = internal.OperationID(opIDRaw)
		opKey := osb.OperationKey(opIDRaw)
		sReq.OperationKey = &opKey
	}

	sResp, err := srv.lastOpGetter.GetLastOperation(r.Context(), osbCtx, &sReq)
	switch {
	case IsNotFoundError(err):
		srv.writeResponse(w, http.StatusGone, map[string]interface{}{})
		return
	case err != nil:
		srv.writeErrorResponse(w, http.StatusBadRequest, err.Error(), "")
		return
	}

	logRespFields := logrus.Fields{
		"action":               "getLastOperation",
		"instance:id":          instanceID,
		"operation:id":         operationID,
		"resp:operation:state": sResp.State,
		"resp:operation:desc":  nil,
	}

	resp := LastOperationSuccessResponseDTO{
		State: internal.OperationState(sResp.State),
	}
	if sResp.Description != nil {
		desc := string(*sResp.Description)
		logRespFields["resp:operation:desc"] = desc
		resp.Description = &desc
	}

	if srv.logger != nil {
		srv.logger.WithFields(logRespFields).Info("action response")
	}
	srv.writeResponse(w, http.StatusOK, resp)
}

func (srv *Server) bindAction(w http.ResponseWriter, r *http.Request) {
	osbCtx, _ := osbContextFromContext(r.Context())

	instanceID := mux.Vars(r)["instance_id"]

	var params BindParametersDTO
	err := httpBodyToDTO(r, &params)
	if err != nil {
		srv.writeErrorResponse(w, http.StatusBadRequest, err.Error(), "cannot get bind parameters from request body")
	}

	err = params.Validate()
	if err != nil {
		srv.writeErrorResponse(w, http.StatusBadRequest, err.Error(), "")
	}

	q := r.URL.Query()
	bindIDRaw := q.Get("binding_id")

	sReq := osb.BindRequest{
		InstanceID: instanceID,
		ServiceID:  params.ServiceID,
		PlanID:     params.PlanID,
		BindingID:  bindIDRaw,
	}
	sResp, err := srv.binder.Bind(r.Context(), osbCtx, &sReq)
	if err != nil {
		srv.writeErrorResponse(w, http.StatusBadRequest, err.Error(), "")
		return
	}

	if srv.logger != nil {
		var keys []string
		for k := range sResp.Credentials {
			keys = append(keys, k)
		}
		logRespFields := logrus.Fields{
			"action":                "bind",
			"resp:async":            false,
			"resp:credentials:keys": keys,
		}
		srv.logger.WithFields(logRespFields).Info("action response")
	}

	egDTO := BindSuccessResponseDTO{
		Credentials: sResp.Credentials,
	}
	srv.writeResponse(w, http.StatusCreated, egDTO)
}

func (srv *Server) unBindAction(w http.ResponseWriter, r *http.Request) {
	srv.writeResponse(w, http.StatusGone, map[string]interface{}{})
}

func (srv *Server) writeResponse(w http.ResponseWriter, code int, object interface{}) {
	writeResponse(w, code, object)
}

func writeResponse(w http.ResponseWriter, code int, object interface{}) {
	data, err := json.Marshal(object)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

func (srv *Server) writeErrorResponse(w http.ResponseWriter, code int, errorMsg, desc string) {
	if srv.logger != nil {
		srv.logger.Warnf("Server responds with error: [HTTP %d]: [%s] [%s]", code, errorMsg, desc)
	}
	writeErrorResponse(w, code, errorMsg, desc)
}

// writeErrorResponse writes error response compatible with OpenServiceBroker API specification.
func writeErrorResponse(w http.ResponseWriter, code int, errorMsg, desc string) {
	dto := struct {
		// Error is a machine readable info on an error.
		// As of 2.13 Open Broker API spec it's NOT passed to entity querying the catalog.
		Error string `json:"error,optional"`

		// Desc is a meaningful error message explaining why the request failed.
		// see: https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#broker-errors
		Desc string `json:"description,optional"`
	}{}

	if errorMsg != "" {
		dto.Error = errorMsg
	}

	if desc != "" {
		dto.Desc = desc
	}
	writeResponse(w, code, &dto)
}

func httpBodyToDTO(r *http.Request, object interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, object)
	if err != nil {
		return err
	}

	return nil
}
