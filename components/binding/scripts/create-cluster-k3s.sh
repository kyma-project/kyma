#!/bin/bash
set -o errexit

cat > registries.yaml <<EOL
mirrors:
registry.localhost:5000:
  endpoint:
  - http://registry.localhost:5000
configs: {}
auths: {}
EOL

sudo mkdir -p /etc/rancher/k3s
sudo cp registries.yaml /etc/rancher/k3s

docker run -d \
-p 5000:5000 \
--restart=always \
--name registry.localhost \
-v $PWD/registry:/var/lib/registry \
registry:2

echo "starting cluster"
curl -sfL https://get.k3s.io | K3S_KUBECONFIG_MODE=777 INSTALL_K3S_EXEC="server --disable traefik" sh -
mkdir -p ~/.kube
cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
chmod 600 ~/.kube/config

host::create_coredns_template(){
cat > coredns-patch.tpl <<EOL
data:
  Corefile: |
    registry.localhost:53 {
        hosts {
            REGISTRY_IP registry.localhost
        }
    }
    .:53 {
        errors
        health
        rewrite name regex (.*)\.local\.kyma\.dev istio-ingressgateway.istio-system.svc.cluster.local
        ready
        kubernetes cluster.local in-addr.arpa ip6.arpa {
          pods insecure
          upstream
          fallthrough in-addr.arpa ip6.arpa
        }
        hosts /etc/coredns/NodeHosts {
          reload 1s
          fallthrough
        }
        prometheus :9153
        forward . /etc/resolv.conf
        cache 30
        loop
        reload
        loadbalance
    }
EOL
}

host::create_coredns_template