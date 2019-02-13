package middlewares

import (
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"
	"net/http"
	"regexp"
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
			httphelpers.RespondWithError(w, apperrors.BadRequest("Client-Certificate-Subject header is empty"))
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

	re := regexp.MustCompile("CN=([^,]+)")
	matches := re.FindStringSubmatch(subject)

	return matches[1]
}
