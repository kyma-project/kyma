#!/usr/bin/env bash -e

CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
WORKING_DIR="${CURRENT_DIR}/tmp-certs"

function command_exists() {
    type "$1" &> /dev/null ;
}

for CMD in openssl kubectl base64 helm; do
	if ! command_exists "${CMD}" ; then
		echo "---> ${CMD} Not found!. Exitting"
		exit 1
	fi
done

mkdir -p "${WORKING_DIR}"

cat <<EOF > "${WORKING_DIR}/openssl.cnf"
[ req ]
#default_bits		= 2048
#default_md		= sha256
#default_keyfile 	= privkey.pem
distinguished_name	= req_distinguished_name
attributes		= req_attributes

[ req_distinguished_name ]
countryName			= Country Name (2 letter code)
countryName_min			= 2
countryName_max			= 2
stateOrProvinceName		= State or Province Name (full name)
localityName			= Locality Name (eg, city)
0.organizationName		= Organization Name (eg, company)
organizationalUnitName		= Organizational Unit Name (eg, section)
commonName			= Common Name (eg, fully qualified host name)
commonName_max			= 64
emailAddress			= Email Address
emailAddress_max		= 64

[ req_attributes ]
challengePassword		= A challenge password
challengePassword_min		= 4
challengePassword_max		= 20
[ v3_ca ]
basicConstraints = critical,CA:TRUE
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer:always
EOF

echo "---> Generate CA"
openssl genrsa -out "${WORKING_DIR}/ca.key.pem" 4096
openssl req -key "${WORKING_DIR}/ca.key.pem" -new -x509 -days 365 -sha256 -out "${WORKING_DIR}/ca.cert.pem" -extensions v3_ca -config "${WORKING_DIR}/openssl.cnf" -subj "/C=PL/ST=Gliwice/L=Gliwice/O=tiller/CN=tiller"

echo "---> Generate Tiller key"
openssl genrsa -out "${WORKING_DIR}/tiller.key.pem" 4096
openssl req -key "${WORKING_DIR}/tiller.key.pem" -new -sha256 -out "${WORKING_DIR}/tiller.csr.pem" -subj "/C=PL/ST=Gliwice/L=Gliwice/O=Tiller Server/CN=tiller-server"
openssl x509 -req -CA "${WORKING_DIR}/ca.cert.pem" -CAkey "${WORKING_DIR}/ca.key.pem" -CAcreateserial -in "${WORKING_DIR}/tiller.csr.pem" -out "${WORKING_DIR}/tiller.cert.pem" -days 365

echo "---> Generate Helm key"
openssl genrsa -out "${WORKING_DIR}/helm.key.pem" 4096
openssl req -key "${WORKING_DIR}/helm.key.pem" -new -sha256 -out "${WORKING_DIR}/helm.csr.pem" -subj "/C=PL/ST=Gliwice/L=Gliwice/O=Helm Client/CN=helm-client"
openssl x509 -req -CA "${WORKING_DIR}/ca.cert.pem" -CAkey "${WORKING_DIR}/ca.key.pem" -CAcreateserial -in "${WORKING_DIR}/helm.csr.pem" -out "${WORKING_DIR}/helm.cert.pem" -days 365

echo "---> Create secrets in k8s"
TILLER_SECRETS=$(cat << EOF
---
apiVersion: v1
data:
  ca.crt: "$(base64 ${WORKING_DIR}/ca.cert.pem | tr -d '\n')"
  tls.crt: "$(base64 ${WORKING_DIR}/tiller.cert.pem | tr -d '\n')"
  tls.key: "$(base64 ${WORKING_DIR}/tiller.key.pem | tr -d '\n')"
kind: Secret
metadata:
  creationTimestamp: null
  labels:
    app: helm
    name: tiller
  name: tiller-secret
  namespace: kube-system
type: Opaque
EOF
)

HELM_SECRETS=$(cat << EOF
---
apiVersion: v1
kind: Namespace
metadata:
  name: kyma-installer
---
apiVersion: v1
data:
  global.helm.ca.crt: "$(base64 ${WORKING_DIR}/ca.cert.pem | tr -d '\n')"
  global.helm.tls.crt: "$(base64 ${WORKING_DIR}/helm.cert.pem | tr -d '\n')"
  global.helm.tls.key: "$(base64 ${WORKING_DIR}/helm.key.pem | tr -d '\n')"
kind: Secret
metadata:
  creationTimestamp: null
  labels:
    installer: overrides
    kyma-project.io/installation: ""
  name: helm-secret
  namespace: kyma-installer
type: Opaque
EOF
)

echo "${TILLER_SECRETS}" | kubectl apply -f -
echo "${HELM_SECRETS}" | kubectl apply -f -

echo "---> Move secrets to helm home"
cp "${WORKING_DIR}/ca.cert.pem" "$(helm home)/ca.pem"
cp "${WORKING_DIR}/helm.cert.pem" "$(helm home)/cert.pem"
cp "${WORKING_DIR}/helm.key.pem" "$(helm home)/key.pem"

echo "---> Cleanup"
rm -rf "${WORKING_DIR}"
