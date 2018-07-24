#!/usr/bin/env bash
  
set -o errexit

FINAL_IMAGE="kyma-on-minikube"
IGNORE_TEST_FAIL=true
RUN_TESTS=true

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --non-interactive)
          NON_INTERACTIVE=true
          shift # past argument
          ;;
        --exit-on-test-fail)
          IGNORE_TEST_FAIL=false
          shift # past argument
          ;;
        --skip-tests)
          RUN_TESTS=false
          shift
          ;;
        *) # unknown option
          POSITIONAL+=("$1") # save it in an array for later
          shift # past argument
          ;;
    esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

DOCKER_RUN_COMMAND="docker run --rm -v /var/lib/docker \
        -p 443:443 \
        -p 8443:8443 \
        -p 8001:8001 \
        -p 9411:9411 \
        -p 32000:32000 \
        -p 32001:32001 \
        -e IGNORE_TEST_FAIL=${IGNORE_TEST_FAIL} \
        -e RUN_TESTS=${RUN_TESTS} \
        --privileged"

if [ -z "${NON_INTERACTIVE}" ]; then
  DOCKER_RUN_COMMAND="${DOCKER_RUN_COMMAND} -it "
fi

DOCKER_RUN_COMMAND="${DOCKER_RUN_COMMAND} ${FINAL_IMAGE}"

if type sudo 1> /dev/null 2> /dev/null;  then
  sudo ${DOCKER_RUN_COMMAND}
else
  ${DOCKER_RUN_COMMAND}
fi
