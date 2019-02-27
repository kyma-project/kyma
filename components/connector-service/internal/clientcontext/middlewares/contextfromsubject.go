package middlewares

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"

	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
)

const (
	subjectSeparator = "\\;"
)

type appContextFromSubjMiddleware struct{}

func NewAppContextFromSubjMiddleware() *appContextFromSubjMiddleware {
	return &appContextFromSubjMiddleware{}
}

func (cc *appContextFromSubjMiddleware) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextExtender, err := prepareContextExtender(r)
		if err != nil {
			httphelpers.RespondWithErrorAndLog(w, apperrors.BadRequest("Invalid certificate subject"))
			return
		}

		reqWithCtx := r.WithContext(contextExtender.ExtendContext(r.Context()))

		handler.ServeHTTP(w, reqWithCtx)
	})
}

func prepareContextExtender(r *http.Request) (clientcontext.ContextExtender, apperrors.AppError) {
	app, group, tenant := parseContextFromSubject(r)
	clusterContext := clientcontext.ClusterContext{
		Group:  group,
		Tenant: tenant,
	}

	if app == "" {
		if clusterContext.IsEmpty() {
			return nil, apperrors.BadRequest("Invalid certificate subject")
		}

		return clusterContext, nil
	}

	return clientcontext.ApplicationContext{
		Application:    app,
		ClusterContext: clusterContext,
	}, nil
}

func parseContextFromSubject(r *http.Request) (application string, group string, tenant string) {
	subject := r.Header.Get(clientcontext.SubjectHeader)
	if subject == "" {
		return "", "", ""
	}

	re := regexp.MustCompile("CN=([^,]+)")
	matches := re.FindStringSubmatch(subject)

	if matches == nil || len(matches) < 2 {
		return "", "", ""
	}

	match := matches[1]

	if strings.Contains(match, subjectSeparator) {
		matchSplitted := strings.Split(match, subjectSeparator)
		return getAt(matchSplitted, 2), getAt(matchSplitted, 1), getAt(matchSplitted, 0)
	}

	return match, "", ""
}

func getAt(slice []string, index int) string {
	if len(slice) <= index {
		return ""
	}

	return slice[index]
}
