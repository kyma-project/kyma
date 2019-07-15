package middlewares

import (
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"

	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
)

type contextFromSubjectExtractor func(subject string) (clientcontext.ContextExtender, apperrors.AppError)

type contextFromSubjMiddleware struct {
	contextFromSubject contextFromSubjectExtractor
	headerParser       certificates.HeaderParser
}

func NewContextFromSubjMiddleware(headerParser certificates.HeaderParser, extractFullContext bool) *contextFromSubjMiddleware {
	var contextFromSubjectExtractor contextFromSubjectExtractor

	if extractFullContext {
		contextFromSubjectExtractor = fullContextFromSubject
	} else {
		contextFromSubjectExtractor = applicationContextFromSubject
	}

	return &contextFromSubjMiddleware{
		contextFromSubject: contextFromSubjectExtractor,
		headerParser:       headerParser,
	}
}

func (cc *contextFromSubjMiddleware) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextExtender, err := cc.parseContextFromSubject(r)
		if err != nil {
			httphelpers.RespondWithErrorAndLog(w, apperrors.BadRequest("Invalid certificate: %s", err.Error()))
			return
		}

		reqWithCtx := r.WithContext(contextExtender.ExtendContext(r.Context()))

		handler.ServeHTTP(w, reqWithCtx)
	})
}

func fullContextFromSubject(subject string) (clientcontext.ContextExtender, apperrors.AppError) {
	tenant := certificates.GetOrganization(subject)
	group := certificates.GetOrganizationalUnit(subject)
	id := certificates.GetCommonName(subject)

	if isAnyEmpty(tenant, group, id) {
		return nil, apperrors.BadRequest("Invalid certificate header, one of the values not provided")
	}

	return clientcontext.ClientContext{
		Group:  group,
		Tenant: tenant,
		ID:     id,
	}, nil
}

func (cc *contextFromSubjMiddleware) parseContextFromSubject(r *http.Request) (clientcontext.ContextExtender, apperrors.AppError) {

	certInfo, e := cc.headerParser.ParseCertificateHeader(*r)

	if e != nil {
		return nil, e
	}

	subject := certInfo.Subject
	if subject == "" {
		return nil, apperrors.BadRequest("Failed to get certificate from header.")
	}

	return cc.contextFromSubject(subject)
}

func applicationContextFromSubject(subject string) (clientcontext.ContextExtender, apperrors.AppError) {
	appName := certificates.GetCommonName(subject)

	if isEmpty(appName) {
		return nil, apperrors.BadRequest("Empty Common Name in certificate header")
	}

	return clientcontext.ClientContext{
		ID: appName,
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

func isEmpty(str string) bool {
	return str == ""
}
