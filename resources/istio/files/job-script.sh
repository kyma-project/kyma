set -e

if [ -f "/etc/istio/overrides.yaml" ]; then
  #New way: just merge default IstioOperator definition with a user-provided one.
  yq merge -x /etc/istio/config.yaml /etc/istio/overrides.yaml > /etc/combo.yaml
  kubectl create cm "${CONFIGMAP_NAME}" -n "${NAMESPACE}" \
    --from-file /etc/istio/config.yaml \
    --from-file /etc/istio/overrides.yaml \
    --from-file /etc/combo.yaml \
    -o yaml --dry-run | kubectl replace -f -
  printf "istioctl manifest apply --wait --readiness-timeout 2m -f /etc/combo.yaml\n"
  istioctl manifest apply --wait --readiness-timeout 2m -f /etc/combo.yaml
else
  #Old way: apply single-value Helm overrides using `istioctl --set "key=val"`
  overrides=$(kubectl get cm --all-namespaces -l "installer=overrides,component=istio" -o go-template --template='{{ range .items }}{{ range $key, $value := .data }}{{ if ne $key "kyma_istio_operator" }}{{ printf "%s: %s\n" $key . }}{{ end }}{{ end }}{{ end }}' )
  overrides_transformed=""

  if [ ! -z "$overrides" ]; then
    while IFS= read -r line; do
      key=$(echo "$line" | cut -d ':' -f 1)
      val=$(echo "$line" | cut -d ':' -f 2 | cut -d ' ' -f 2)

      case $key in
        pilot.resources* )
          new_key=$(echo "$key" | cut -d '.' -f 2-)
          key=$(echo "trafficManagement.components.pilot.k8s.$new_key")
          ;;
        mixer.loadshedding.mode* )
          key=$(echo "values.mixer.telemetry.loadshedding.mode")
          ;;
        mixer.telemetry.resources* )
          new_key=$(echo "$key" | cut -d '.' -f 3-)
          key=$(echo "telemetry.components.telemetry.k8s.$new_key")
          ;;
        mixer.policy.resources* )
          new_key=$(echo "$key" | cut -d '.' -f 3-)
          key=$(echo "policy.components.policy.k8s.$new_key")
          ;;
        gateways.istio-ingressgateway.resources* )
          new_key=$(echo "$key" | cut -d '.' -f 3-)
          key=$(echo "gateways.components.ingressGateway.k8s.$new_key")
          ;;
        gateways.istio-ingressgateway.autoscaleMin* )
          key=$(echo "gateways.components.ingressGateway.k8s.hpaSpec.minReplicas")
          ;;
        gateways.istio-ingressgateway.autoscaleMax*)
          key=$(echo "gateways.components.ingressGateway.k8s.hpaSpec.maxReplicas")
          ;;
        * )
          key=$(echo "values.$key")
          ;;
      esac

      if [ -z "$val" ]; then
        val=$(echo '""')
      fi
      overrides_transformed=$(printf "$overrides_transformed --set $key=$val")
    done <<< "$overrides"
  fi

  printf "istioctl manifest apply --wait --readiness-timeout 2m -f /etc/istio/config.yaml ${overrides_transformed}\n"
  istioctl manifest apply --wait --readiness-timeout 2m -f /etc/istio/config.yaml ${overrides_transformed}
fi

echo "Apply custom kyma manifests"
kubectl apply -f /etc/manifests

echo "Apply Kyma related checks and patches"
kubectl patch MutatingWebhookConfiguration istio-sidecar-injector --type 'json' -p '[{"op":"add","path":"/webhooks/0/namespaceSelector/matchExpressions/-","value":{"key":"gardener.cloud/purpose","operator":"NotIn","values":["kube-system"]}}]'
