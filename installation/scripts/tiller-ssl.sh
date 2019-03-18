#!/usr/bin/env bash -e

CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

function command_exists() {
    type "$1" &> /dev/null ;
}

for CMD in openssl kubectl base64 helm; do
	if ! command_exists "${CMD}" ; then
		echo "---> ${CMD} Not found!. Exitting"
		exit 1
	fi
done

cat <<EOF > ./openssl.cnf
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
openssl genrsa -out ./ca.key.pem 4096
openssl req -key ca.key.pem -new -x509 -days 365 -sha256 -out ca.cert.pem -extensions v3_ca -config ./openssl.cnf -subj "/C=PL/ST=Gliwice/L=Gliwice/O=tiller/CN=tiller"

echo "---> Generate Tiller key"
openssl genrsa -out ./tiller.key.pem 4096
openssl req -key tiller.key.pem -new -sha256 -out tiller.csr.pem -subj "/C=PL/ST=Gliwice/L=Gliwice/O=Tiller Server/CN=tiller-server"
openssl x509 -req -CA ca.cert.pem -CAkey ca.key.pem -CAcreateserial -in tiller.csr.pem -out tiller.cert.pem -days 365

echo "---> Generate Helm key"
openssl genrsa -out ./helm.key.pem 4096
openssl req -key helm.key.pem -new -sha256 -out helm.csr.pem -subj "/C=PL/ST=Gliwice/L=Gliwice/O=Helm Client/CN=helm-client"
openssl x509 -req -CA ca.cert.pem -CAkey ca.key.pem -CAcreateserial -in helm.csr.pem -out helm.cert.pem  -days 365

echo "---> Create secrets in k8s"
TILLER_SECRETS=$(cat << EOF
---
apiVersion: v1
data:
  ca.crt: "$(base64 ./ca.cert.pem)"
  tls.crt: "$(base64 ./tiller.cert.pem)"
  tls.key: "$(base64 ./tiller.key.pem)"
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
data:
  ca.crt: "$(base64 ./ca.cert.pem)"
  tls.crt: "$(base64 ./helm.cert.pem)"
  tls.key: "$(base64 ./helm.key.pem)"
kind: Secret
metadata:
  creationTimestamp: null
  labels:
    app: helm
    name: helm
  name: helm-secret
  namespace: kyma-system
type: Opaque
EOF
)

echo "${TILLER_SECRETS}" | kubectl apply -f -

echo "---> Move secrets to helm home"
mv ca.cert.pem $(helm home)/ca.pem
mv helm.cert.pem $(helm home)/cert.pem
mv helm.key.pem $(helm home)/key.pem

echo "---> Cleanup"
rm ./openssl.cnf
rm ./*.pem
rm ./*.srl

