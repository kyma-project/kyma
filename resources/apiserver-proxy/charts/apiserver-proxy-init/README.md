# API Server Proxy Init

## Overview

This chart generates certificates required by API Server Proxy.

## Details

Service `apiserver-proxy-ssl` is a LoadBalancer. Job, depending on the domain configuration, generates certificates for given IP address  and saves them in the `apiserver-proxy-tls-cert` secret, which is later mounted to API Server Proxy deployment.