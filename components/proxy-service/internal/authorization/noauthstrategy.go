package authorization

import (
	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	"net/http"
)

func newNoAuthStrategy() noAuthStrategy{
	return noAuthStrategy{}
}

type noAuthStrategy struct {
}

func (ns noAuthStrategy) Setup(r *http.Request) apperrors.AppError {
	return nil
}

func (ns noAuthStrategy) Reset() {

}
