param (
    [string]$VM_DRIVER = "hyperv",
    [string]$DOMAIN = "kyma.local"
)

$CURRENT_DIR = Split-Path $MyInvocation.MyCommand.Path
$KUBERNETES_VERSION = "1.10.0"

Write-Output @"
################################################################################
# Minikube setup with driver ${VM_DRIVER} and kubernetes ${KUBERNETES_VERSION}
################################################################################
"@

function CheckIfMinikubeIsInitialized() {
    $cmd = "minikube status --format '{{.MinikubeStatus}}'"
    $minikubeStatus = (Invoke-Expression -Command $cmd) | Out-String
    
    if ($minikubeStatus -ne "") {
        Write-Output "Minikube is already initialized"
        $deleteMinikube = Read-Host "Do you want to remove previous minikube cluster [y/N]"
        if ($deleteMinikube -eq "y") {
            $cmd = "minikube delete"
            Invoke-Expression -Command $cmd
        }
        else {
            Write-Output "Starting minikube cancelled"
            exit 1
        }
    }
}

function InitializeMinikubeConfig () {
    $cmd = "minikube config unset ingress"
    Invoke-Expression -Command $cmd
}

function UploadDexTlsCertForApiserver() {
    Write-Output "Parsing DEX TLS certificate"

    $yaml = Get-Content -Path "${CURRENT_DIR}\..\resources\local-tls-certs.yaml" | Select-String -Pattern 'tls.crt: .*'
    $yaml = $yaml.ToString().Replace("tls.crt:", "").Trim()
    $yaml = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($yaml))
    $cert = $yaml.ToString()
    
    Write-Output "Saving DEX TLS certificate in the container file."

    $crtDir = "${HOME}\.minikube\files\dex"
    $crtFile = "${crtDir}\dex-ca.crt"
    
    if (!(Test-Path -Path $crtDir)) {
        New-Item -ItemType Directory -Path $crtDir | Out-Null
    }

    $cert | Set-Content $crtFile
}

function StartMinikube() {
    $cmd = "minikube start"`
        + " --memory 8192"`
        + " --cpus 4"`
        + " --extra-config=apiserver.Authorization.Mode=RBAC"`
        + " --extra-config=apiserver.Authentication.OIDC.IssuerURL='https://dex.${DOMAIN}'"`
        + " --extra-config=apiserver.Authentication.OIDC.CAFile=/home/docker/dex/dex-ca.crt"`
        + " --extra-config=apiserver.Authentication.OIDC.ClientID=kyma-client"`
        + " --extra-config=apiserver.Authentication.OIDC.UsernameClaim=email"`
        + " --extra-config=apiserver.Authentication.OIDC.GroupsClaim=groups"`
        + " --extra-config=apiserver.GenericServerRunOptions.CorsAllowedOriginList='.*'"`
        + " --extra-config=controller-manager.ClusterSigningCertFile='/var/lib/localkube/certs/ca.crt'"`
        + " --extra-config=controller-manager.ClusterSigningKeyFile='/var/lib/localkube/certs/ca.key'"`
        + " --extra-config=apiserver.Admission.PluginNames='Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota,PodPreset,PersistentVolumeLabel'"`
        + " --kubernetes-version=v${KUBERNETES_VERSION}"`
        + " --feature-gates='MountPropagation=false'"`
        + " --vm-driver=${VM_DRIVER}"`
        + " -b=localkube"

    if ($VM_DRIVER -eq "hyperv") {
        $cmd += " --hyperv-virtual-switch='${env.HYPERV_VIRTUAL_SW}'"
    }
    
    Invoke-Expression -Command $cmd
}

function WaitForMinikubeToBeUp() {
    Write-Output "Waiting for minikube to be up..."

    $limit = 15
    $counter = 0
    $clusterStatus = ""
    
    while ($counter -lt $limit) {
        $counter += 1
        Write-Output "Keep calm, there are ${limit} possibilities and so far it is attempt number ${counter}"
      
        $cmd = "minikube status --format '{{.MinikubeStatus}}'"
        $clusterStatus = (Invoke-Expression -Command $cmd) | Out-String
        $clusterStatus = $clusterStatus.Trim()
        if ($clusterStatus -eq "Running") {
            break
        }
      
        Start-Sleep -Seconds 1
    }
}

function AddDevDomainsToEtcHosts([string[]]$hostnamesPrefixes) {
    $hostnames = $hostnamesPrefixes | % {"$_.${DOMAIN}"}
    $cmd = "minikube ip"
    $minikubeIp = (Invoke-Expression -Command $cmd | Out-String).Trim()

    Write-Output "Minikube IP address: ${minikubeIp}"

    $cmd = "minikube ssh 'echo `"127.0.0.1 ${hostnames}`" | sudo tee -a /etc/hosts'"
    Invoke-Expression -Command $cmd

    $winHostsPath = "C:\Windows\system32\drivers\etc\hosts"
    Get-Content -Path $winHostsPath | Select-String -Pattern $DOMAIN -NotMatch | Out-File -FilePath $winHostsPath

    "${minikubeIp} ${hostnames}" | Out-File $winHostsPath -Append
}

CheckIfMinikubeIsInitialized
InitializeMinikubeConfig
UploadDexTlsCertForApiserver
StartMinikube
WaitForMinikubeToBeUp
AddDevDomainsToEtcHosts "apiserver", "console", "catalog", "instances", "dex", "docs", "lambdas-ui", "ui-api" "minio", "jaeger", "grafana", "configurations-generator"
