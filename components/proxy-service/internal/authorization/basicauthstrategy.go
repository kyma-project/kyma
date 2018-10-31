/*
 * [y] hybris Platform
 *
 * Copyright (c) 2000-2018 hybris AG
 * All rights reserved.
 *
 * This software is the confidential and proprietary information of hybris
 * ("Confidential Information"). You shall not disclose such Confidential
 * Information and shall use it only in accordance with the terms of the
 * license agreement you entered into with hybris.
 */
package authorization

import (
	"net/http/httputil"
	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	"net/http"
	"encoding/base64"
	"github.com/kyma-project/kyma/components/proxy-service/internal/httpconsts"
	"fmt"
)

type basicAuthStrategy struct {
	username     string
	password string
}

func newBasicAuthStrategy(username, password string) basicAuthStrategy{
	return basicAuthStrategy {
		username: username,
		password: password,
	}
}

func (o basicAuthStrategy) Setup(proxy *httputil.ReverseProxy, r *http.Request) apperrors.AppError {
	r.Header.Set(httpconsts.HeaderAuthorization, fmt.Sprintf("Basic %s", basicAuth(o.username, o.password)))

	return nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password

	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (o basicAuthStrategy) Reset() {
}