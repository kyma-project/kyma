apiVersion: v1
kind: Secret
metadata:
  name: application-connector-certificate-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    kyma-project.io/installation: ""
type: Opaque
data:
  global.applicationConnectorCa: ""
  global.applicationConnectorCaKey: ""
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cluster-certificate-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    kyma-project.io/installation: ""
data:
  global.tlsCrt: "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUU4RENDQXRpZ0F3SUJBZ0lKQUpBKzlrQzZZZnZlTUEwR0NTcUdTSWIzRFFFQkN3VUFNQmN4RlRBVEJnTlYKQkFNTURDb3VhM2x0WVM1c2IyTmhiREFlRncweE9EQTNNVGd3T0RNNE1UTmFGdzB5T0RBM01UVXdPRE00TVROYQpNQmN4RlRBVEJnTlZCQU1NRENvdWEzbHRZUzVzYjJOaGJEQ0NBaUl3RFFZSktvWklodmNOQVFFQkJRQURnZ0lQCkFEQ0NBZ29DZ2dJQkFPZ05XbVROOU1ranlrUjdvQ0JGSFVqL01EcWh5bml3NEJITGo4ZTdDTFV5dFdwVHZXTkoKU1FiaFZDK3c5NkhHbU50MHZGUTR4OExUa3NNUmorcVZrdkcwKzBDTE1WQm14UjBVdnpZR09QRVRKREtsNTZkcQppMEM1Y2dnU3dkNjcveWxRZmQxc3FEVVBHM0pVZlNOSFFRWSs1SkUwaUpjZXhCQ3cvS200UXlUQXB0aEwzdEgxCm1ZRFBBQ0hUdVpsbGc2RVN4Z2RKY20rVmg5UkRvbWlON250ZjBZVG1xV1VIZXhZUkUrcTFGY0VhUHg5L2Q5QUwKZWd1WXZHTkduMVA4K1F3ZE5DKzVaMEVGa1EyS3RtalRBR09La2xJUG5NQkZubXhGSEtNZzNTa3RkYzYzSU5TNwpjSDM3dDNjb2l4N09HcDllcnRvSFZ5K2U5KzdiYnI3Z0lhWHNORVZUcjRGWXJIMlNPeHI1MjVRQUpHTys1dllQCkNYSlJJdXFNZWU3ZTg3aDRIU3JFVVdWTnhOdElBR1ZNYlRtV2F6aDJsSEFoazJjZVIzUkRaQzB3NGFOd3NRU1UKZE1yMmFCaUFobmFNSG40YnZIZnd1OHFRcjF5aVRUNG00SkltRkZTNDN2bGNETDZmVUJnNnF0U2p6QjF0WTFjTwp2Mnl1QXQzR1lKSENVMFd4R1ZEQk80T0ZncEc3aE5SRHoyRzVFMXQvRWU2VDUxdys2K295eXhib01pK09kVWdNCnBZOGlzcDdBTWhtdlZFOUdtNUhwQkpBRjYvcmE1WUNnNDh1SWtybkdZbm85eVNKcGhLU0JsMFRML01PNVFRaUoKb0hFNnV5ZkJXQjlZZ1NBdDl0MjJqU3FSWkpSdEtrbHB6bFJIRkROWTRwZlRraXlPSlFXYlM4RXJBZ01CQUFHagpQekE5TUFrR0ExVWRFd1FDTUFBd0N3WURWUjBQQkFRREFnWGdNQ01HQTFVZEVRUWNNQnFDQ210NWJXRXViRzlqCllXeUNEQ291YTNsdFlTNXNiMk5oYkRBTkJna3Foa2lHOXcwQkFRc0ZBQU9DQWdFQUlYYTlwenlyQ2dzMTRTOHUKZVFZdkorNEFzUE9uT1RGcExkaVl5UkVyNXdyNmJuMXUvMjZxc2FKckpxbkkyRk16SmdEQVRwZEtjbXRHYjBUOQp3S2wrYUJHcFFKcThrUWJwakVGTHhaWDJzaUNrRG82WittaUcrRjRKMHpKa3BKK0JHMS92eGZKbk0zK1ptdXQ5Ck9RV2ZjYTN3UHlhTWRDbGIyZjQwYlRFaFo5Mk9kcWlQMzFMbDlHWExSZmhaNTNsUzF0QWdvUGZoR25NbFY4b2MKWmxuSUROK25wS0Nma2tXUDJZUjlRLy9pa01tM01YRm9RSFppaVJseVZHSGFKZWRLMmNOQzlUYk4xNDFTaWZHZQo3V2FsQVBNcWNOQ3F3YStnN2RFSmR3ZjlRMklJTml0SjlDUVprT1dUZElYY2VHK2lZWWUrQXpmK1NkaHBocVdPCllFcDF6ek40dXI5U2VxU3NSaU9WY0RzVUFSa1M0clgrb0Vzb2hHL1Q5OTcrSDhjR2gzczl6TE84emtwRXZKSmEKS05QT0N5ODhVeEFOV2RRejFLMXRKVVQ2c3hkd0FEcXRJQnNPemhYVjlybDRRNStlZExlcmZPcUtCbUFRMUY5Swo2L1l0ZlNyY0JpeXZEU24wdFJ3OHJLRFVQU1hFNDFldXArOURNeThLVGl6T0RPTXVMSnR2dkJrTEFpNGNYQjVBCjQxMjBEdHdZQXNyNzNZYVl2SW8rWjV2OGZ4TjF3M3IwYS9KOVhZQlg3S3p1OFl4MnNUNWtWM2dNTHFCTXBaa3gKY29FTjNSandDMmV4VHl6dGc1ak1ZN2U4VFJ4OFFTeUxkK0pBd2t1Tm01NlNkcHFHNTE3cktJYkVMNDZzbkd0UgpCYUVOK01GeXNqdDU3ejhKQXJDMzhBMFN5dTQ9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K"
  global.tlsKey: "LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JSUpRd0lCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQ1Mwd2dna3BBZ0VBQW9JQ0FRRG9EVnBremZUSkk4cEUKZTZBZ1JSMUkvekE2b2NwNHNPQVJ5NC9IdXdpMU1yVnFVNzFqU1VrRzRWUXZzUGVoeHBqYmRMeFVPTWZDMDVMRApFWS9xbFpMeHRQdEFpekZRWnNVZEZMODJCamp4RXlReXBlZW5hb3RBdVhJSUVzSGV1LzhwVUgzZGJLZzFEeHR5ClZIMGpSMEVHUHVTUk5JaVhIc1FRc1B5cHVFTWt3S2JZUzk3UjlabUF6d0FoMDdtWlpZT2hFc1lIU1hKdmxZZlUKUTZKb2plNTdYOUdFNXFsbEIzc1dFUlBxdFJYQkdqOGZmM2ZRQzNvTG1MeGpScDlUL1BrTUhUUXZ1V2RCQlpFTgppclpvMHdCamlwSlNENXpBUlo1c1JSeWpJTjBwTFhYT3R5RFV1M0I5KzdkM0tJc2V6aHFmWHE3YUIxY3ZudmZ1CjIyNis0Q0dsN0RSRlU2K0JXS3g5a2pzYStkdVVBQ1JqdnViMkR3bHlVU0xxakhudTN2TzRlQjBxeEZGbFRjVGIKU0FCbFRHMDVsbXM0ZHBSd0laTm5Ia2QwUTJRdE1PR2pjTEVFbEhUSzltZ1lnSVoyakI1K0c3eDM4THZLa0s5YwpvazArSnVDU0poUlV1Tjc1WEF5K24xQVlPcXJVbzh3ZGJXTlhEcjlzcmdMZHhtQ1J3bE5Gc1JsUXdUdURoWUtSCnU0VFVRODlodVJOYmZ4SHVrK2RjUHV2cU1zc1c2REl2am5WSURLV1BJcktld0RJWnIxUlBScHVSNlFTUUJldjYKMnVXQW9PUExpSks1eG1KNlBja2lhWVNrZ1pkRXkvekR1VUVJaWFCeE9yc253VmdmV0lFZ0xmYmR0bzBxa1dTVQpiU3BKYWM1VVJ4UXpXT0tYMDVJc2ppVUZtMHZCS3dJREFRQUJBb0lDQUZDN3ZKUlB0M2QzVlRyb1MvaU9NemNmCldaYzhqT1hhbThwMUtRdlRQWjlWQ2hyNUVXNEdwRHFaa0tHYkR6eWdqTFBsZEZSVkFPTCtteFAwK3o0aFZlTjAKRk9vS3cxaDJ1T042UVdBNVgvdzNyYU5WWnpndThFM1BkeVhwNkx0bWFzcmo3elpuUkVwWmZESVZ4UWZPRllobgp2enZwckEvdnEwVW5YbkJwNUNwWVFIUUdTWHFBMlN3Z1dLcHNNQ2wzVVFsc0w2dC9XU29MT3h1VmdGNmg2clBQCnpXUlFuK1MvYW9wdDNLRU82WWVxYXdXNVltVG1hVXE1aytseU82S0w0OVhjSHpqdlowWU8rcjFjWWtRc0RQbVUKejMxdll4amQzOVZKWWtJNi85Y0Fzdmo5YTVXM3ROYVFDZStTRW56Z05oRDJieHo1NnRKdG0xTGwweXpqYTdEVgp0OW9aQVd6WURMOTVnMDFZWVlQMG51c3A4WWlNOXd0OUtLc2NkcUd3RFN5WHNrWFBrcWFLVjNVcHJYdmhFbFVaCkErMmtjcm9VaDNGbEV4bGpwUmtKVkhOZXJ3NHRLRi9oYTFWRjZPdE10eTVQcXV0N0dGQmIvamtWeUg5cnpueWUKTXQyTWVyTTVPazMwd1NuTThISUdTUXpxYlplekJEZlNaUzRzcWdZZnBIMlhtMEs0SjgrRUowQ2hhMXZVSmVNMAoyZ284d00vaHljdmtqTEgxSmM3OEhpaVBTQ01udkpHemUxc2tWdmtRRFhBSFdldzBTUHpUSTZHYjZCb0Y4aVNECm0wZjR2azNoV3NlUWZBaXVZSnlUeUZXNmRhOGE1K2lpSDN4cVRsUUN1MDN1Nmo0U0l4aThJZlNmd0YwQTBldVAKNGtzalZTZVZyT3ExUnlvNUtpR3hBb0lCQVFEOWZtYnl6aW9QdVhRYzl0QXBxMUpSMzErQzlCdFFzcDg5WkZkSQpQaU5xaTJ3NVlVcTA0OFM2Z3VBb3JGOHNObUI3QjhWa1JlclowQ3hub2NHY0tleWdTYWsvME5qVElndk5weGJwCnBGbkFnRjlmbW1oTEl2SlF2REo2Q0ZidDRCQlZIdkJEWlYyQnZqK0k3NUxkK01jN2RPVDdFek1FRjBXcUdzY2MKTUpyNjRXQi9UMkF5dWR1YXlRT2NobmJFQ25FUmdRcHFlbG54MVBraytqbGNvYUs5QjFYUStVOUgzOHppM0FYNApENUxMY0Nhem9YYWlvS0swckNlNE5Ga2hOVXd0TFV0QXhSTXk2aksyZUZudWczUFRlY3N1WktNMElITktqZ0dCCnpGanZVb2tMcFVFb3BJa3FHM09yc0xmanpHZW9jaVFPUXNEdzlUb3lXL0FSOFhmWkFvSUJBUURxV0s2TThVN3EKUXJPeTYzNnpEZlBaZ2ljeXlsUWVoOFdMclBlbW5NeWdQYzR4eWoybnMrTmVRSnNEUmtPT2tWY250SEFYaTcwWgoyT0NCV3dwZHJuTXlSc3RIMU05bjdFNU5TVHZlZDlkU3YxUzRBb3NzS1hDSmgyUHBjYjV0OE9nL3ZGTlNYUlUyClk2aUorWTdOcDBZNDNxSlJOVnlRemd3YmFzaEpiUVdkVFFoVEVhdEVRS2JsUlZSblhlQWRjOXlhNUpHbkRpaTkKbFQrRWEzdFpvN1dha05oeHJkYjVuTkZ3a0xoNEs3ZkFtT3ZzMVBMQWx0SUZqeURCeDEvY1ZHblpDUDBVQmJqZgpkU2FueXBBdVRuMzd1VUwrcXpPVDlYWEZENllGT0x0NWV4d3RxdnUwSzZCNjcvajFFTlJDRk45RnMzWlV5RFFXClZUaDcybFhWU3NLakFvSUJBUUR6ZE5pdXpTNDhWK0thaHJpNXJGNnRYeGkrRG0vRmV5ZlFzSFBiWUVKbmEyd1AKVjgrR0YxS3p4a28vQmYySjJ0ZWlrWDRVcGNtK1UxNnlVUG8vWDB4eFRRMk55cWpUYmRsa005dWZuVWJOeVB6UQpOdDEvZkJxNVMyWTNLWmREY25SOUsrK1k2dHQ1Wmh4akNhUkdKMDVCWGkwa3JmWExNZ2FvTG51WUtWNVBJUEdxCms3TlNSSW9UQ0llOVpxN2Q3U0ZXckZZeW1UdVZOUFByZlo1bHhwOGphTTRVbTd4MnpReGJ2UERHb3o1YXdHV0wKRThGNncwaEF1UzZValVJazBLbE9vamVxQnh3L1JBcGNrUTNlTXNXbEQwNENTb2tyNFJhWlBmVllrY2ZBWWNaWgpOdWR6ZjBKMC9GU0ZTbjN4L0RoNTROV2NGS1IxUnpBVGVaVUJ4cVZSQW9JQkFRRFlDNmZvVWpOQnJ2ckNFVzkrCkhYZlk1Ni9CbUZ4U3hUTHU0U2h6VncwakViZTltVWljQ2pDc1hQMUwySVJCdEdaWU9YWTVqdDlvSzlSV0RSdVMKWUZqZFdmemduU1lWRmZyZUw0emRQVGlxbGEvQjhMNWptVlNoeGNycmxheE02Uk1FWjFlZGtDa1ZPbTFQdmwzVAo1TW5OZGhySXFWeE1OMWxjRVdiU29vclJpUW9Lb3poMHRQSG9YckZBbG9BZVJ3bHpWeE9jb21ZVzJiaDBHUzdmCjVoaHZoZWUxYmVISnY3UXFoWkU3WUhxSU9iTVBaUWJqWEdnRkxmMnlDRitzM2Jtem1DRFJTN0V6ZVdxSXVDdVMKTlZUYU0rSzZyQlRoN0NLRjZUWlNqQW55SmZoRmRlT1ZKNzlNZDEzYWVJaG0zNTB6UWc3dWZKL2drdkorNUR2TApacC9uQW9JQkFBVlg4WHpFTzdMVk1sbENKZFVUdURTdXNPcDNrQVlFZ2dZNFFRM3FNTlcxRnl5WEM3WjBGOWJFCmtTSEhkalJtU2RUbFZueGN2UW1KTS9WL2tJanpNUHhFT3NCS1BVVkR3N3BhOHdiejlGcTRPOCtJb3lqN1ZXclcKMmExL1FNWXlzSGlpTlBzNy8vWUtvMy9rdkhCWUY2SnNkenkyQkVSTkQ0aTlVOWhDN0RqcGxKR3BSNktMTVBsegpNWFJ3VjVTM2V3cnBXZVcxQW5ONC9EKy9zUGlNQTNnS0swSlBFdGVBV1dndEZHTnNBSkJnaFBoUExxQi9CcDUvCkhOeC96M0w0MWtqRnpqOHNWaHMrVDRZYlhiaGF2R2xxc2h5ZldQbnRhV1VOMG15MjU1RFdQUDhWa24yeFNlV2kKVm1hVW5TSDBTZ2tlUENMRnlra25yQzgxU2pXZkRBMD0KLS0tLS1FTkQgUFJJVkFURSBLRVktLS0tLQo="
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: installation-config-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    kyma-project.io/installation: ""
data:
  global.isLocalEnv: "true"
  global.domainName: "kyma.local"
  global.adminPassword: ""
  global.minikubeIP: ""
  nginx-ingress.controller.service.loadBalancerIP: ""
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: istio
    kyma-project.io/installation: ""
data:
  gateways.istio-ingressgateway.loadBalancerIP: ""
  gateways.istio-ingressgateway.type: "NodePort"
  gateways.istio-ingressgateway.autoscaleEnabled: "false"

  pilot.resources.limits.memory: 1024Mi
  pilot.resources.limits.cpu: 500m
  pilot.resources.requests.memory: 512Mi
  pilot.resources.requests.cpu: 250m
  pilot.autoscaleEnabled: "false"

  mixer.policy.resources.limits.memory: 2048Mi
  mixer.policy.resources.limits.cpu: 500m
  mixer.policy.resources.requests.memory: 512Mi
  mixer.policy.resources.requests.cpu: 300m

  mixer.telemetry.resources.limits.memory: 2048Mi
  mixer.telemetry.resources.limits.cpu: 500m
  mixer.telemetry.resources.requests.memory: 512Mi
  mixer.telemetry.resources.requests.cpu: 300m
  mixer.loadshedding.mode: disabled

  mixer.policy.autoscaleEnabled: "false"
  mixer.telemetry.autoscaleEnabled: "false"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: service-catalog-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: service-catalog
    kyma-project.io/installation: ""
data:
  etcd-stateful.etcd.resources.limits.memory: 256Mi
  etcd-stateful.replicaCount: "1"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: helm-broker-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: helm-broker
    kyma-project.io/installation: ""
data:
  global.isDevelopMode: "true" # global, because subcharts also use it
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: dex-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: dex
    kyma-project.io/installation: ""
data:
  telemetry.enabled: "false"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: application-connector-tests
  namespace: kyma-installer
  labels:
    installer: overrides
    component: application-connector
    kyma-project.io/installation: ""
data:
  application-operator.tests.enabled: "false"
  application-registry.tests.enabled: "false"
  connector-service.tests.enabled: "false"
  tests.application_connector_tests.enabled: "false"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: compass-runtime-agent-tests
  namespace: kyma-installer
  labels:
    installer: overrides
    component: compass-runtime-agent
    kyma-project.io/installation: ""
data:
  compassRuntimeAgent.tests.enabled: "false"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: core-tests
  namespace: kyma-installer
  labels:
    installer: overrides
    component: core
    kyma-project.io/installation: ""
data:
  kubeless.tests.enabled: "false"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-plane-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: istio
    installerOnly: "true"
    kyma-project.io/installation: ""
data:
  kyma_istio_control_plane: |-
    apiVersion: install.istio.io/v1alpha2
    kind: IstioControlPlane
    spec:
      autoInjection:
        components:
          injector:
            enabled: true
            k8s:
              affinity:
                podAntiAffinity:
                  preferredDuringSchedulingIgnoredDuringExecution: []
                  requiredDuringSchedulingIgnoredDuringExecution: []
              nodeSelector: {}
              replicaCount: 1
              strategy:
                rollingUpdate:
                  maxSurge: 100%
                  maxUnavailable: 25%
              tolerations: []
        enabled: true
      cni:
        components:
          cni:
            enabled: false
        enabled: false
      configManagement:
        components:
          galley:
            enabled: true
            k8s:
              affinity:
                podAntiAffinity:
                  preferredDuringSchedulingIgnoredDuringExecution: []
                  requiredDuringSchedulingIgnoredDuringExecution: []
              nodeSelector: {}
              replicaCount: 1
              strategy:
                rollingUpdate:
                  maxSurge: 100%
                  maxUnavailable: 25%
              tolerations: []
        enabled: true
      coreDNS:
        components:
          coreDNS:
            enabled: false
        enabled: false
      gateways:
        components:
          egressGateway:
            enabled: false
          ingressGateway:
            enabled: true
            k8s:
              affinity:
                podAntiAffinity:
                  preferredDuringSchedulingIgnoredDuringExecution: []
                  requiredDuringSchedulingIgnoredDuringExecution: []
              env:
                - name: ISTIO_META_ROUTER_MODE
                  value: sni-dnat
              nodeSelector: {}
              podAnnotations: {}
              resources:
                limits:
                  cpu: 2000m
                  memory: 256Mi
                requests:
                  cpu: 100m
                  memory: 96Mi
              strategy:
                rollingUpdate:
                  maxSurge: 1
                  maxUnavailable: 0
              tolerations: []
              overlays:
                - kind: Deployment
                  name: istio-ingressgateway
                  patches:
                    - path: spec.template.spec.containers.[name:istio-proxy].ports.[containerPort:80].hostPort
                      value: 80
                    - path: spec.template.spec.containers.[name:istio-proxy].ports.[containerPort:443].hostPort
                      value: 443
      policy:
        components:
          policy:
            enabled: true
            k8s:
              replicaCount: 1
              resources:
                limits:
                  cpu: 500m
                  memory: 2048Mi
                requests:
                  cpu: 300m
                  memory: 512Mi
              strategy:
                rollingUpdate:
                  maxSurge: 1
                  maxUnavailable: 0
        enabled: true
      security:
        components:
          certManager:
            enabled: false
          citadel:
            enabled: true
            k8s:
              affinity:
                podAntiAffinity:
                  preferredDuringSchedulingIgnoredDuringExecution: []
                  requiredDuringSchedulingIgnoredDuringExecution: []
              env: []
              nodeSelector: {}
              replicaCount: 1
              strategy:
                rollingUpdate:
                  maxSurge: 100%
                  maxUnavailable: 25%
              tolerations: []
          nodeAgent:
            enabled: false
        enabled: true
      telemetry:
        components:
          telemetry:
            enabled: true
            k8s:
              replicaCount: 1
              resources:
                limits:
                  cpu: 500m
                  memory: 2048Mi
                requests:
                  cpu: 300m
                  memory: 512Mi
              strategy:
                rollingUpdate:
                  maxSurge: 1
                  maxUnavailable: 0
        enabled: true
      trafficManagement:
        components:
          pilot:
            enabled: true
            k8s:
              affinity:
                podAntiAffinity:
                  preferredDuringSchedulingIgnoredDuringExecution: []
                  requiredDuringSchedulingIgnoredDuringExecution: []
              env:
                - name: PADU
                  value: padu-minikube
                - name: GODEBUG
                  value: gctrace=1
                - name: PILOT_HTTP10
                  value: "1"
                - name: PILOT_PUSH_THROTTLE
                  value: "100"
              nodeSelector: {}
              resources:
                limits:
                  cpu: 500m
                  memory: 1024Mi
                requests:
                  cpu: 250m
                  memory: 512Mi
              strategy:
                rollingUpdate:
                  maxSurge: 1
                  maxUnavailable: 0
              tolerations: []
        enabled: true
      values:
        global:
          controlPlaneSecurityEnabled: true
          mtls:
            enabled: true
          policyCheckFailOpen: true
          proxy:
            resources:
              requests:
                cpu: 10m
                memory: 10Mi
              limits:
                cpu: 100m
                memory: 50Mi
        galley:
          image: galley
        gateways:
          istio-egressgateway:
            autoscaleEnabled: true
            autoscaleMax: 5
            autoscaleMin: 1
            cpu:
              targetAverageUtilization: 80
            env:
              iSTIO_META_ROUTER_MODE: sni-dnat
            labels:
              app: istio-egressgateway
              istio: egressgateway
            ports:
              - name: http2
                port: 80
              - name: https
                port: 443
              - name: tls
                port: 15443
                targetPort: 15443
            resources:
              limits:
                cpu: 2000m
                memory: 1024Mi
              requests:
                cpu: 100m
                memory: 128Mi
            secretVolumes:
              - mountPath: /etc/istio/egressgateway-certs
                name: egressgateway-certs
                secretName: istio-egressgateway-certs
              - mountPath: /etc/istio/egressgateway-ca-certs
                name: egressgateway-ca-certs
                secretName: istio-egressgateway-ca-certs
            type: ClusterIP
          istio-ingressgateway:
            applicationPorts: ""
            autoscaleEnabled: false
            cpu:
              targetAverageUtilization: 80
            externalIPs: []
            labels:
              app: istio-ingressgateway
              istio: ingressgateway
            loadBalancerIP: ""
            loadBalancerSourceRanges: []
            meshExpansionPorts:
              - name: tcp-pilot-grpc-tls
                port: 15011
                targetPort: 15011
              - name: tcp-mixer-grpc-tls
                port: 15004
                targetPort: 15004
              - name: tcp-citadel-grpc-tls
                port: 8060
                targetPort: 8060
              - name: tcp-dns-tls
                port: 853
                targetPort: 853
            ports:
              - name: status-port
                port: 15020
                targetPort: 15020
              - name: http2
                nodePort: 31380
                port: 80
                targetPort: 80
              - name: https
                nodePort: 31390
                port: 443
              - name: https-kiali
                port: 15029
                targetPort: 15029
              - name: https-prometheus
                port: 15030
                targetPort: 15030
              - name: https-grafana
                port: 15031
                targetPort: 15031
              - name: https-tracing
                port: 15032
                targetPort: 15032
              - name: tls
                port: 15443
                targetPort: 15443
            sds:
              enabled: true
              image: node-agent-k8s
              resources:
                limits:
                  cpu: 50m
                  memory: 64Mi
                requests:
                  cpu: 10m
                  memory: 16Mi
            secretVolumes:
              - mountPath: /etc/istio/ingressgateway-certs
                name: ingressgateway-certs
                secretName: istio-ingressgateway-certs
              - mountPath: /etc/istio/ingressgateway-ca-certs
                name: ingressgateway-ca-certs
                secretName: istio-ingressgateway-ca-certs
            type: NodePort
        grafana:
          enabled: false
        kiali:
          enabled: false
        mixer:
          adapters:
            kubernetesenv:
              enabled: true
            prometheus:
              enabled: true
              metricsExpiryDuration: 10m
            stdio:
              enabled: false
              outputAsJson: true
            useAdapterCRDs: false
          policy:
            autoscaleEnabled: false
            cpu:
              targetAverageUtilization: 80
            enabled: true
          telemetry:
            autoscaleEnabled: false
            cpu:
              targetAverageUtilization: 80
            reportBatchMaxEntries: 100
            reportBatchMaxTime: 1s
            sessionAffinityEnabled: false
        pilot:
          autoscaleEnabled: false
          cpu:
            targetAverageUtilization: 80
          enableProtocolSniffingForInbound: false
          enableProtocolSniffingForOutbound: true
          image: pilot
          keepaliveMaxServerConnectionAge: 30m
          policy:
            enabled: true
          sidecar: true
          traceSampling: 1
        prometheus:
          enabled: false
        security:
          citadelHealthCheck: false
          createMeshPolicy: true
          enableNamespacesByDefault: true
          image: citadel
          selfSigned: true
          workloadCertTtl: 2160h
        sidecarInjectorWebhook:
          alwaysInjectSelector: []
          enableNamespacesByDefault: true
          image: sidecar_injector
          neverInjectSelector: []
          rewriteAppHTTPProbe: true
        tracing:
          enabled: false
---
