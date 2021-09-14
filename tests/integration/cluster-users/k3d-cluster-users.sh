#!/usr/bin/env bash

set -e

KYMA_PROJECT_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )"/../../../ &> /dev/null && pwd )"

function host::os() {
  local host_os
  case "$(uname -s)" in
  Linux)
    host_os=linux
    ;;
  *)
    echo >&2 -e "Unsupported host OS. Must be Linux"
    exit 1
    ;;
  esac
  echo "${host_os}"
}

function install::docker() {
  if ! command -v docker &>/dev/null; then
    curl -fsSL https://download.docker.com/linux/debian/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
    echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian \
    $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list >/dev/null
    sudo apt-get update
    sudo apt-get install docker-ce docker-ce-cli containerd.io -y
  fi

}

function install::docker_compose() {
  if ! command -v docker-compose &>/dev/null; then
    sudo curl -L "https://github.com/docker/compose/releases/download/1.29.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    sudo chmod +x /usr/local/bin/docker-compose
  fi

}

function install::kyma_cli() {
  local settings
  local kyma_version
  mkdir -p "/usr/local/bin"
  local os
  os=$(host::os)

  pushd "/usr/local/bin" || exit

  echo "Install kyma CLI ${os} locally to /usr/local/bin..."

  curl -sSLo kyma "https://storage.googleapis.com/kyma-cli-stable/kyma-${os}?alt=media"
  chmod +x kyma
  kyma_version=$(kyma version --client)
  echo "Kyma CLI version: ${kyma_version}"

  echo "OK"

  popd || exit

  eval "${settings}"
}

function install::add_kyma_cert_to_truststore() {
  cat kyma-cert.pem | sudo tee -a /etc/ssl/certs/ca-certificates.crt
}

function install::golang() {
  if ! command -v go &>/dev/null; then
    wget --quiet https://golang.org/dl/go1.17.linux-amd64.tar.gz
    tar -C /usr/local -xzf go1.17.linux-amd64.tar.gz
    export PATH=$PATH:/usr/local/go/bin
  fi

}

function install::k3d() {
  curl -s https://raw.githubusercontent.com/rancher/k3d/main/install.sh | bash
}

function install::dependencies() {
  install::golang
  install::docker
  install::docker_compose
  install::kyma_cli
  install::add_kyma_cert_to_truststore
}

function run::docker_compose() {
  docker-compose up --build --detach
}

function hydra::create_client() {
  run::docker_compose
  sleep 8 #wait for hydra
  docker-compose exec -T hydra \
    hydra clients create \
    --endpoint http://127.0.0.1:4445 \
    --id auth-code-client \
    --secret secret \
    --grant-types implicit \
    --response-types id_token,code,token \
    --scope openid,read,write \
    --callbacks http://testclient3.example.com
}

function provision::kyma_k3d() {
  kyma provision k3d --k3d-arg '-v' --k3d-arg '/dev/mapper:/dev/mapper' \
    --k3d-arg '--k3s-server-arg' --k3d-arg '--kube-apiserver-arg=oidc-issuer-url=https://oauth2-fake.local.kyma.dev/' \
    --k3d-arg '--k3s-server-arg' --k3d-arg '-kube-apiserver-arg=oidc-username-claim=email' \
    --k3d-arg '--k3s-server-arg' --k3d-arg '--kube-apiserver-arg=oidc-groups-claim=groups' \
    --k3d-arg '--k3s-server-arg' --k3d-arg '--kube-apiserver-arg=oidc-client-id=auth-code-client' \
    --k3d-arg '--k3s-server-arg' --k3d-arg '--kube-apiserver-arg=v=9' \
    --k3d-arg '-v' --k3d-arg '/etc/ssl/certs/ca-certificates.crt:/etc/pki/tls/certs/ca-bundle.crt'
}

function install::kyma_cluster_users() {
  yes | kyma deploy --components-file ./components.yaml
}

function cluster-users::hydra::setup_env_vars() {

  local default_password="1234"

  echo "--> Setting up variables"
  ADMIN_EMAIL="admin@kyma.cx"
  echo "---> ADMIN_EMAIL: $ADMIN_EMAIL"
  export ADMIN_EMAIL
  ADMIN_PASSWORD=$default_password
  echo "---> ADMIN_PASSWORD: $ADMIN_PASSWORD"
  export ADMIN_PASSWORD

  DEVELOPER_EMAIL="developer@kyma.cx"
  echo "---> DEVELOPER_EMAIL: $DEVELOPER_EMAIL"
  export DEVELOPER_EMAIL
  DEVELOPER_PASSWORD=$default_password
  echo "---> DEVELOPER_PASSWORD: $DEVELOPER_PASSWORD"
  export DEVELOPER_PASSWORD

  VIEW_EMAIL="read-only-user@kyma.cx"
  echo "---> VIEW_EMAIL: $VIEW_EMAIL"
  export VIEW_EMAIL
  VIEW_PASSWORD=$default_password
  echo "---> VIEW_PASSWORD: $VIEW_PASSWORD"
  export VIEW_PASSWORD

  NAMESPACE_ADMIN_EMAIL="namespace.admin@kyma.cx"
  echo "---> NAMESPACE_ADMIN_EMAIL: $NAMESPACE_ADMIN_EMAIL"
  export NAMESPACE_ADMIN_EMAIL
  NAMESPACE_ADMIN_PASSWORD=$default_password
  echo "---> NAMESPACE_ADMIN_PASSWORD: $NAMESPACE_ADMIN_PASSWORD"
  export NAMESPACE_ADMIN_PASSWORD

  export KYMA_SYSTEM="kyma-system"
  export SYSTEM_NAMESPACE="kyma-system"
  # shellcheck disable=SC2155
  local namespace_id=$(tr </dev/urandom -dc 'a-zA-Z0-9' | fold -w 5 | head -n 1 | tr '[:upper:]' '[:lower:]')
  CUSTOM_NAMESPACE="test-namespace-$namespace_id"
  echo "---> CUSTOM_NAMESPACE: $CUSTOM_NAMESPACE"
  export CUSTOM_NAMESPACE
  export NAMESPACE="default"
}

function run::cluster-users-tests() {
  cluster-users::hydra::setup_env_vars
  pushd $KYMA_PROJECT_ROOT
  OIDC_PROVIDER=hydra bash resources/cluster-users/files/sar-test.sh
}

install::dependencies
hydra::create_client
provision::kyma_k3d
install::kyma_cluster_users
run::cluster-users-tests
