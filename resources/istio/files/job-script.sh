
overrides=$(kubectl get cm --all-namespaces -l "installer=overrides,component=istio,installerOnly!=true" -o go-template --template='{{ range .items }}{{ range $key, $value := .data }}{{ printf "%s: %s\n" $key . }}{{ end }}{{ end }}' )
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
    overrides_transformed=$(printf "$overrides_transformed --set \"$key=$val\"")
  done <<< "$overrides"
fi

printf "istioctl manifest apply -f /etc/istio/config.yaml ${overrides_transformed}\n"
istioctl manifest apply -f /etc/istio/config.yaml ${overrides_transformed}

while [ "$(kubectl get po -n istio-system -l app=sidecarInjectorWebhook -o jsonpath='{ .items[0].status.phase}')" != "Running" ]
do
  echo "sidecar injector still not running. Waiting..."
  sleep 1s
done
echo "sidecar injector is running"
echo "patching api-server destination rule"
kubectl patch destinationrules.networking.istio.io -n istio-system api-server --type merge --patch '{"spec": {"trafficPolicy": { "connectionPool" : { "tcp": {"connectTimeout": "30s"}}}}}'
