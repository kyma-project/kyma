#!/usr/bin/env bash

WORKING_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'
export APP_CONNECTOR_CERT_DIR="$(mktemp -d -t cert.XXXXXX)"

check_dependencies() {
  # Dependencies
  echo "Checking dependencies:"

  local missing=0

  ## Curl
  $( which curl &> /dev/null; )
  if [[ $? != 0 ]]
    then
      echo -e "${RED}curl${NC}"
      missing=1
    else echo -e "${GREEN}curl${NC}"
  fi

  ## OpenSSL
  $( which openssl &> /dev/null; )
  if [[ $? != 0 ]]
    then
      echo -e "${RED}openssl${NC}"
      missing=1
    else echo -e "${GREEN}openssl${NC}"
  fi

  ## JQ
  $( which jq &> /dev/null; )
  if [[ $? != 0 ]]
    then
      echo -e "${RED}jq${NC}"
      missing=1
    else echo -e "${GREEN}jq${NC}"
  fi

  ## Base64
  $( which base64 &> /dev/null; )
  if [[ $? != 0 ]]
    then
      echo -e "${RED}base64${NC}"
      missing=1
    else echo -e "${GREEN}base64${NC}"
  fi

  ## Kubectl
  $( which kubectl &> /dev/null; )
  if [[ $? != 0 ]]
    then
      echo -e "${RED}kubectl${NC}"
      missing=1
    else echo -e "${GREEN}kubectl${NC}"
  fi

  if [[ $missing != 0 ]]
    then
      echo ""
      echo -e "${RED}Dependencies missing, aborting${NC}"
      exit 1;
  fi

  echo ""
  echo -e "${GREEN}Dependencies fulfilled${NC}"
  echo ""
}

echo "One-Click-Integration script started with following parameters:"
echo -e "Key file: \t\t ${k}"
echo -e "Connector-Service URL: \t ${u}"
echo ""
echo -e $(kubectl config current-context)

check_dependencies

kubectl create ns app-gateway-test
kubectl apply -f ${WORKING_DIR}/application-gateway.yaml -n app-gateway-test

echo -e "${GREEN}Kubeconfig file present, fetching url${NC}"
echo "Creating TokenRequest"
printf "apiVersion: applicationconnector.kyma-project.io/v1alpha1\nkind: TokenRequest\nmetadata:\n  name: perf-app" > ${APP_CONNECTOR_CERT_DIR}/generated.yaml

# Create a TokenRequest and temporarily save it to the file
kcapply=$( kubectl apply -f ${APP_CONNECTOR_CERT_DIR}/generated.yaml -n app-gateway-test )
if [[ $? != 0 ]]
  then
    echo $kcapply
    echo -e "${RED}TokenRequest creation failed${NC}"
    $( rm ${APP_CONNECTOR_CERT_DIR}/generated.yaml )
    exit 1;
fi

echo "Polling CR for CSR url"
i="0"
while [ $i -lt 6 ]
do
  # Give controller some time to write the token to the CR
  sleep $i;
  tokenRequest=$( kubectl -n app-gateway-test get TokenRequest perf-app -o=jsonpath='{.status.url}' )
  if [[ ! -z "${tokenRequest}" ]]
    then
      break;
  else
    if [[ $1 -eq 5 ]]
      then
        echo -e "${RED}Maximum number of retries exceeded, exitting${NC}"
        exit 1;
    fi
  fi

   i=$[ $i + 1 ]

done
# Overwrite url with the fetched one
u=$tokenRequest
echo -e "${GREEN}Polling finished!${NC}"
echo ""


## GET /info
echo "Fetching info for CSR"
infoRequest=$( curl -k -L "${u}" )
if [[ $? != 0 ]]
  then
    echo -e "${RED}Request Failed${NC}"
    exit 1;
fi

csrUrl=$( echo ${infoRequest} | jq -r '.csrUrl')
subject=$( echo ${infoRequest} | jq -r '.certificate.subject')

### Check if csr url and subject values were returned
if [[ "$csrUrl" == null ]] || [[ "$subject" == null ]]; then
  echo -e "${RED}Info request failed${NC}"
  echo -e "${RED}Status code: \t $( echo ${infoRequest} | jq '.code' )${NC}"
  echo -e "${RED}Error: \t\t $( echo ${infoRequest} | jq '.error' )${NC}"
  exit 1;
fi

### Fix subject (if old version of connector-service is used)
subject=$( echo "${subject}" | tr , / )

### Check if subject string starts with '/' and fix if it does not
if [[ $subject != /* ]]; then subject="/$subject"; fi
echo ""
echo -e "${GREEN}Info request succeeded${NC}"
echo -e "${GREEN}csrUrl: \t ${csrUrl}${NC}"
echo -e "${GREEN}subject: \t ${subject}${NC}"

## If key was not provided create it
if [[ -z "${k}" ]]; then
  echo ""
  echo -e "${GREEN}Creating key file${NC}"
  echo ""
  key=$( openssl genrsa -out ${APP_CONNECTOR_CERT_DIR}/generated.key 4096 )
  if [[ $? != 0 ]]
    then
      echo -e "${RED}Failed to create a key file${NC}"
      exit 1;
  fi

  echo ""
  echo -e "${GREEN}Creating CSR${NC}"
  echo ""
  csr=$( openssl req -new -out ${APP_CONNECTOR_CERT_DIR}/generated.csr -key ${APP_CONNECTOR_CERT_DIR}/generated.key -subj "${subject}" )
else
  echo ""
  echo -e "${GREEN}Creating CSR${NC}"
  echo ""
  csr=$( openssl req -new -out ${APP_CONNECTOR_CERT_DIR}/generated.csr -key "${k}" -subj "${subject}" )
fi

if [[ $? != 0 ]]
  then
    echo -e "${RED}Failed to create CSR${NC}"
    exit 1;
fi

## Base64 encode the CSR
csrb64=$( base64 ${APP_CONNECTOR_CERT_DIR}/generated.csr | tr -d "\n" )
if [[ $? != 0 ]]
  then
    echo -e "${RED}CSR not found${NC}"
    exit 1;
fi

## Send POST request to client-certs
echo -e "${GREEN}Sending CSR${NC}"
signRequest=$( curl -k -L -X POST -H "Content-Type: application/json" -d '{"csr":"'"${csrb64}"'"}' "${csrUrl}" )
if [[ $? != 0 ]]
  then
    echo -e "${RED}Request Failed${NC}"
    exit 1;
fi

crtb64=$( echo "${signRequest}" | jq -r '.crt' )
if [[ "$crtb64" == null ]]; then
  echo -e "${RED}Sign request failed${NC}"
  echo -e "${RED}Status code: \t $( echo ${signRequest} | jq '.code' )${NC}"
  echo -e "${RED}Error: \t\t $( echo ${signRequest} | jq '.error' )${NC}"
  exit 1;
fi

echo ""
echo -e "${GREEN}Sign request succeeded${NC}"
echo ""

## Decode response and save as a certificate file
echo "Decoding and saving certificate"
echo ""
crt=$( echo "${crtb64}" | openssl enc -base64 -d -A > ${APP_CONNECTOR_CERT_DIR}/generated.crt )
if [[ $? != 0 ]]
  then
    echo -e "${RED}Decoding failed${NC}"
    exit 1;
fi

echo -e "${GREEN}Certificate retrieved successfully${NC}"
echo ""
echo -e "${GREEN}Creating pem file${NC}"
cat ${APP_CONNECTOR_CERT_DIR}/generated.key ${APP_CONNECTOR_CERT_DIR}/generated.crt > ${APP_CONNECTOR_CERT_DIR}/generated.pem

echo ""
echo "Cleaning up..."
$( rm ${APP_CONNECTOR_CERT_DIR}/generated.csr )
$( rm ${APP_CONNECTOR_CERT_DIR}/generated.yaml )

$( kubectl -n app-gateway-test delete TokenRequest perf-app &> /dev/null; )


echo ""
echo -e "${GREEN}Done!${NC}"