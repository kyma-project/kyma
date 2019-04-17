package middlewares

import (
	"net/http"
	"regexp"

	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"

	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
)

type contextFromSubjectExtractor func(subject string) (clientcontext.ContextExtender, apperrors.AppError)

type contextFromSubjMiddleware struct {
	contextFromSubject contextFromSubjectExtractor
}

func NewContextFromSubjMiddleware(extractFullContext bool) *contextFromSubjMiddleware {
	var contextFromSubjectExtractor contextFromSubjectExtractor

	if extractFullContext {
		contextFromSubjectExtractor = fullContextFromSubject
	} else {
		contextFromSubjectExtractor = applicationContextFromSubject
	}

	return &contextFromSubjMiddleware{
		contextFromSubject: contextFromSubjectExtractor,
	}
}

func (cc *contextFromSubjMiddleware) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextExtender, err := cc.parseContextFromSubject(r)
		if err != nil {
			httphelpers.RespondWithErrorAndLog(w, apperrors.BadRequest("Invalid certificate subject"))
			return
		}

		reqWithCtx := r.WithContext(contextExtender.ExtendContext(r.Context()))

		handler.ServeHTTP(w, reqWithCtx)
	})
}

func (cc *contextFromSubjMiddleware) parseContextFromSubject(r *http.Request) (clientcontext.ContextExtender, apperrors.AppError) {
	subject := r.Header.Get(clientcontext.SubjectHeader)
	if subject == "" {
		return nil, apperrors.BadRequest("Failed to get certificate subject from header.")
	}

	return cc.contextFromSubject(subject)
}

func applicationContextFromSubject(subject string) (clientcontext.ContextExtender, apperrors.AppError) {
	appName := getCommonName(subject)

	if isEmpty(appName) {
		return nil, apperrors.BadRequest("Empty Common Name in subject header")
	}

	return clientcontext.ApplicationContext{
		Application:    appName,
		ClusterContext: clientcontext.ClusterContext{},
	}, nil
}

func fullContextFromSubject(subject string) (clientcontext.ContextExtender, apperrors.AppError) {
	tenant := getOrganization(subject)
	group := getOrganizationalUnit(subject)
	commonName := getCommonName(subject)

	if isAnyEmpty(tenant, group, commonName) {
		return nil, apperrors.BadRequest("Invalid subject header, one of the values not provided")
	}

	clusterContext := clientcontext.ClusterContext{
		Group:  group,
		Tenant: tenant,
	}

	if commonName == clientcontext.RuntimeDefaultCommonName {
		return clusterContext, nil
	}

	return clientcontext.ApplicationContext{
		Application:    commonName,
		ClusterContext: clusterContext,
	}, nil
}

func isAnyEmpty(str ...string) bool {
	for _, s := range str {
		if isEmpty(s) {
			return true
		}
	}

	return false
}

func getCommonName(subject string) string {
	return getRegexMatch("CN=([^,]+)", subject)
}

func getOrganization(subject string) string {
	return getRegexMatch("O=([^,]+)", subject)
}

func getOrganizationalUnit(subject string) string {
	return getRegexMatch("OU=([^,]+)", subject)
}

func getRegexMatch(regex, text string) string {
	cnRegex := regexp.MustCompile(regex)
	matches := cnRegex.FindStringSubmatch(text)

	if matches == nil || len(matches) < 2 {
		return ""
	}

	return matches[1]
}

func isEmpty(str string) bool {
	return str == ""
}
