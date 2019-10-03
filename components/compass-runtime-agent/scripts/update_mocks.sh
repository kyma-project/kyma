#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

go generate ${GOPATH}/src/github.com/kyma-project/kyma/components/compass-runtime-agent/internal/...
