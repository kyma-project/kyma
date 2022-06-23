package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const (
	USER  = "user"
	PASS  = "password"
	TOKEN = "538509cc-a60a-4f4f-b2e6-a852cbc2b965"
	CERT  = "/etc/secret-volume/server.crt"
	KEY   = "/etc/secret-volume/server.key"
	CA    = "/etc/secret-volume/ca.crt"
)

func main() {
	e := echo.New()

	caCert, err := ioutil.ReadFile(CA)
	if err != nil {
		e.Logger.Fatal(err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		ClientCAs:  caCertPool,
		ClientAuth: tls.VerifyClientCertIfGiven,
	}

	server := http.Server{
		Addr:      ":8080",
		Handler:   e,
		TLSConfig: tlsConfig,
	}

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/v1/unprotected", alwaysOK)
	e.GET("/v1/health", alwaysOK)
	e.GET("/v1/basic", alwaysOK, basicAuth(USER, PASS))
	e.GET("/v1/oauth/ok", alwaysOK, oauth(TOKEN))
	e.POST("/v1/oauth/token", token(USER, PASS))
	e.GET("/v1/mtlsoauth/ok", alwaysOK, oauth(TOKEN))
	e.POST("/v1/mtlsoauth/token", mtlsToken(USER))

	if err := server.ListenAndServeTLS(CERT, KEY); err != http.ErrServerClosed {
		e.Logger.Fatal(err)
	}
}

func alwaysOK(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func mtlsToken(clientID string) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Request().TLS == nil {
			return c.String(http.StatusUnauthorized, "Missing certificate")
		}

		id := c.FormValue("client_id")

		if clientID != id {
			c.Logger().Warn("Incorrect oauth data: id=", id)
			return c.String(http.StatusForbidden, "Invalid mTLS OAuth")
		}

		type OauthResponse struct {
			AccessToken string `json:"access_token"`
			TokenType   string `json:"token_type"`
			ExpiresIn   int    `json:"expires_in"`
			Scope       string `json:"scope"`
		}

		response := OauthResponse{AccessToken: TOKEN, TokenType: "bearer", ExpiresIn: 3600, Scope: "basic"}

		return c.JSON(http.StatusOK, response)
	}
}

func token(clientID, clientSecret string) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.FormValue("client_id")
		secret := c.FormValue("client_secret")
		if clientID != id || clientSecret != secret {
			c.Logger().Warn("Incorrect oauth data: id=", id, "secret=", secret)
			return c.String(http.StatusForbidden, "Invalid OAuth")
		}

		type OauthResponse struct {
			AccessToken string `json:"access_token"`
			TokenType   string `json:"token_type"`
			ExpiresIn   int    `json:"expires_in"`
			Scope       string `json:"scope"`
		}

		response := OauthResponse{AccessToken: TOKEN, TokenType: "bearer", ExpiresIn: 3600, Scope: "basic"}

		return c.JSON(http.StatusOK, response)
	}
}

func oauth(token string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			t := c.Request().Header.Get("Authorization")
			if len(t) == 0 {
				c.Logger().Warn("Missing Authorization header")
				return c.NoContent(http.StatusForbidden)
			}

			if !strings.HasPrefix(t, "Bearer ") {
				c.Logger().Warn("Authorization header is missing 'Bearer '")
				return c.NoContent(http.StatusForbidden)
			}

			t = strings.TrimPrefix(t, "Bearer ")

			if t != token {
				c.Logger().Warn("Invalid oauth token:", t)
				return c.NoContent(http.StatusForbidden)
			}

			return next(c)
		}
	}
}

func basicAuth(username, password string) echo.MiddlewareFunc {
	return middleware.BasicAuth(func(u, p string, c echo.Context) (bool, error) {
		if u != username || p != password {
			c.Logger().Warn("Invalid basic auth:", u, ":", p)
			return false, nil
		}
		return true, nil
	})
}
