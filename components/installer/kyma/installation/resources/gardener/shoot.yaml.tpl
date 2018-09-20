kind: Shoot
apiVersion: garden.sapcloud.io/v1beta1
metadata:
  name: __CLUSTER_NAME__
  namespace: garden-__PROJECT_NAME__
  annotations:
    garden.sapcloud.io/purpose: __PURPOSE__
spec:
  addons:
    cluster-autoscaler:
      enabled: true
    heapster:
      enabled: true
    kubernetes-dashboard:
      enabled: true
    nginx-ingress:
      enabled: true
    monocular:
      enabled: false
  backup:
    schedule: '*/5 * * * *'
    maximum: 7
  cloud:
    profile: az
    region: westeurope
    secretBindingRef:
      name: __NAME_OF_AZURE_SECRET__
    azure:
      networks:
        vnet:
          cidr: 10.250.0.0/16
        workers: 10.250.0.0/19
      workers:
        - name: kyma-worker
          machineType: Standard_DS2_v2
          autoScalerMin: 3
          autoScalerMax: 4
          volumeType: standard
          volumeSize: 50Gi
  dns:
    provider: aws-route53
    domain: __CLUSTER_NAME__.__PROJECT_NAME__.shoot.canary.k8s-hana.ondemand.com
  kubernetes:
    allowPrivilegedContainers: true
    kubeAPIServer:
      runtimeConfig:
        settings.k8s.io/v1alpha1: true
      oidcConfig:
        caBundle: |
          __TLS_CERT__
        clientID: kyma-client
        groupsClaim: groups
        issuerURL: 'https://dex.__DOMAIN__'
        usernameClaim: email
    version: 1.10.5
  maintenance:
    autoUpdate:
      kubernetesVersion: true
    timeWindow:
      begin: 010000+0000
      end: 020000+0000