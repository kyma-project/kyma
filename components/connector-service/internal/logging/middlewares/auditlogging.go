package middlewares

import (
	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"
	"github.com/sirupsen/logrus"
	"net/http"
)

type AuditLogMessages struct {
	startingOperationMsg string
	operationSuccessfulMsg string
	operationFailedMsg string
}

type auditlogggingmiddleware struct {
	connectorClientExtractor clientcontext.ConnectorClientExtractor
	auditLogMessages AuditLogMessages
}

func NewAuditLoggingMiddleware(connectorClientExtractor clientcontext.ConnectorClientExtractor, auditLogMessages AuditLogMessages) *auditlogggingmiddleware {
	return &auditlogggingmiddleware{
		connectorClientExtractor: connectorClientExtractor,
		auditLogMessages: auditLogMessages,
	}
}

func (alm *auditlogggingmiddleware) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextLogger := alm.getContextLogger(r)

		contextLogger.Info(alm.auditLogMessages.startingOperationMsg)
		writerWithStatus := httphelpers.WriterWithStatus{ResponseWriter: w}

		handler.ServeHTTP(&writerWithStatus, r)

		if writerWithStatus.IsSuccessful() {
			contextLogger.Info(alm.auditLogMessages.operationSuccessfulMsg)
		} else {
			contextLogger.Info(alm.auditLogMessages.operationFailedMsg)
		}
	})
}

func (alm *auditlogggingmiddleware) getContextLogger(r *http.Request) *logrus.Entry {
	contextService, err := alm.connectorClientExtractor(r.Context())

	if err == nil {
		return contextService.GetLogger()
	}

	return logrus.NewEntry(logrus.StandardLogger())
}


