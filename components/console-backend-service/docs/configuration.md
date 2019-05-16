# Configuration
This document describes configuration details of the application.

## Environmental Variables
Use the following environment variables to configure the application:

| Name | Required | Default | Description |
|------|----------|---------|-------------|
| APP_HOST | No | `127.0.0.1` | The host on which the HTTP server listens. |
| APP_PORT | No | `3000` | The port on which the HTTP server listens. |
| APP_ALLOWED_ORIGINS | No | `*` | Origins that have access to the HTTP server. Origins must be comma-separated list of strings. |
| APP_SERVER_TIMEOUT | No | `10s` | The period of time after which the system kills active requests and stops the server. |
| APP_VERBOSE | No | No | Show detailed logs in the application. |
| APP_KUBECONFIG_PATH | No |  | The path to the `kubeconfig` file, needed for running an application outside of the cluster. |
| APP_INFORMER_RESYNC_PERIOD | No | `10m` | The period of time after which the system resynchronizes the informers. |
| APP_CONTENT_ADDRESS | No | `minio.kyma.local` | The address of the content storage server. |
| APP_CONTENT_PORT | No | `443` | The port on which the content storage server listens. |
| APP_CONTENT_ACCESS_KEY | Yes |  | The access key required to sign in to the content storage server. |
| APP_CONTENT_SECRET_KEY | Yes |  | The secret key required to sign in to the content storage server. |
| APP_CONTENT_BUCKET | No | `content` | The name of the bucket with the content. |
| APP_CONTENT_SECURE | No | `true` | Use HTTPS for the connection with the content storage server. |
| APP_CONTENT_EXTERNAL_ADDRESS | No |  | The external address of the content storage server. If not set, the system uses the `APP_CONTENT_ADDRESS` variable. |
| APP_CONTENT_ASSETS_FOLDER | No | `assets` | The name of the `assets` folder. |
| APP_CONTENT_VERIFY_SSL | No | `true` | Ignore invalid SSL certificates. |
| APP_APPLICATION_GATEWAY_STATUS_REFRESH_PERIOD | No | `15s` | The period of time after which the application refreshes the Application statuses. |
| APP_APPLICATION_GATEWAY_STATUS_CALL_TIMEOUT | No | `500ms` | The timeout of the HTTP call status check. |
| APP_APPLICATION_GATEWAY_INTEGRATION_NAMESPACE | Yes |  | The namespace with gateway services. |
| APP_APPLICATION_CONNECTOR_URL | Yes |  | The address of the connector service. |
| APP_APPLICATION_CONNECTOR_CALL_HTTP_TIMEOUT | No | `500ms` | The timeout of the HTTP call. |
| APP_APPLICATION_VERIFY_SSL | No | `true` | Ignore invalid SSL certificates. |
| APP_OIDC_ISSUER_URL | Yes | | The URL of the OpenID issuer. Used to verify the OIDC JWT. |
| APP_OIDC_CLIENT_ID | Yes | | The client ID for which the JWT must be issued for. |
| APP_OIDC_CA_FILE | No | host's root CA set | One of the Certificate Authorities defined in this file is used to verify the OpenID server's certificate. |
| APP_OIDC_USERNAME_CLAIM | No | `email` | The OpenID claim to use as the user name. |
| APP_OIDC_USERNAME_PREFIX | No | | If provided, all usernames will be prefixed with this value. |
| APP_OIDC_GROUPS_CLAIM | No | `groups` | The name of a custom OpenID Connect claim for specifying user groups. |
| APP_OIDC_GROUPS_PREFIX | NO | | If provided, all groups will be prefixed with this value. |
| APP_OIDC_SUPPORTED_SIGNING_ALGS | NO | `RS256` | Comma-separated list of the allowed JOSE asymmetric signing algorithms. Tokens with the 'alg' header value not in this list will be rejected. Values are defined by [RFC 7518](https://tools.ietf.org/html/rfc7518#section-3.1). |

## Configure logger verbosity level
This application uses `glog` to log messages. Pass command line arguments described in the [glog.go](https://github.com/golang/glog/blob/master/glog.go) document to customize the log, such as log level and output.

For example:
```bash
go run main.go --stderrthreshold=INFO -logtostderr=false
```
