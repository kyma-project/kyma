package middlewares

import (
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"
	"github.com/sirupsen/logrus"
)

type AuditLogMessages struct {
	StartingOperationMsg   string
	OperationSuccessfulMsg string
	OperationFailedMsg     string
}

type auditloggingmiddleware struct {
	connectorClientExtractor clientcontext.ConnectorClientExtractor
	auditLogMessages         AuditLogMessages
}

func NewAuditLoggingMiddleware(connectorClientExtractor clientcontext.ConnectorClientExtractor, auditLogMessages AuditLogMessages) *auditloggingmiddleware {
	return &auditloggingmiddleware{
		connectorClientExtractor: connectorClientExtractor,
		auditLogMessages:         auditLogMessages,
	}
}

func (alm *auditloggingmiddleware) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextLogger := alm.getContextLogger(r)

		contextLogger.Info(alm.auditLogMessages.StartingOperationMsg)
		writerWithStatus := httphelpers.WriterWithStatus{ResponseWriter: w}

		handler.ServeHTTP(&writerWithStatus, r)

		if writerWithStatus.IsSuccessful() {
			contextLogger.Info(alm.auditLogMessages.OperationSuccessfulMsg)
		} else {
			contextLogger.Info(alm.auditLogMessages.OperationFailedMsg)
		}
	})
}

func (alm *auditloggingmiddleware) getContextLogger(r *http.Request) *logrus.Entry {
	contextService, err := alm.connectorClientExtractor(r.Context())

	if err == nil {
		return contextService.GetLogger()
	}

	return logrus.NewEntry(logrus.StandardLogger())
}
