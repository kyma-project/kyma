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
	"net/http"
	"net/http/httputil"
	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	log "github.com/sirupsen/logrus"
	"github.com/kyma-project/kyma/components/proxy-service/internal/httpconsts"
)


type oauthStrategy struct {
	oauthClient OAuthClient
	clientId     string
	clientSecret string
	url          string
}

func newOAuthStrategy(oauthClient OAuthClient, clientId, clientSecret, url string) oauthStrategy{
	return oauthStrategy {
		oauthClient: oauthClient,
		clientId: clientId,
		clientSecret: clientSecret,
		url: url,
	}
}

func (o oauthStrategy) Setup(proxy *httputil.ReverseProxy, r *http.Request) apperrors.AppError {
	token, err := o.oauthClient.GetToken(o.clientId, o.clientSecret, o.url)
	if err != nil {
		log.Errorf("failed to get token : '%s'", err)
		return err
	}

	r.Header.Set(httpconsts.HeaderAuthorization, token)
	log.Infof("OAuth token fetched. Adding Authorization header: %s", r.Header.Get(httpconsts.HeaderAuthorization))

	return nil
}

func (o oauthStrategy) Reset() {
	o.oauthClient.InvalidateTokenCache(o.clientId)
}

