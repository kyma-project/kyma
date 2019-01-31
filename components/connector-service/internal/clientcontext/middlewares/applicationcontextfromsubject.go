package middlewares

import (
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"
	"net/http"
	"strings"
)

type appContextFromSubjMiddleware struct{}

func NewAppContextFromSubjMiddleware() *appContextFromSubjMiddleware {
	return &appContextFromSubjMiddleware{}
}

func (cc *appContextFromSubjMiddleware) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appContext := clientcontext.ApplicationContext{
			Application: extractApplicationFromSubject(r),
		}

		if appContext.Application == "" {
			httphelpers.RespondWithError(w, apperrors.BadRequest("Application context is empty"))
			return
		}

		reqWithCtx := r.WithContext(appContext.ExtendContext(r.Context()))

		handler.ServeHTTP(w, reqWithCtx)
	})
}

func extractApplicationFromSubject(r *http.Request) string {
	subject := r.Header.Get(clientcontext.SubjectHeader)

	if subject == "" {
		return ""
	}

	index := strings.LastIndex(subject, "CN")

	return subject[index+3:]
}
