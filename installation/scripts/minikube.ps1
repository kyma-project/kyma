param (
    [string]$VM_DRIVER = "hyperv",
    [string]$DOMAIN = "kyma.local",
    [string]$DISK_SIZE = "20g",
    [string]$MEMORY = "8196"
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

    $cmd = "minikube addons enable heapster"
    Invoke-Expression -Command $cmd
}

function StartMinikube() {
    $cmd = "minikube start"`
        + " --memory ${MEMORY}"`
        + " --cpus 4"`
        + " --extra-config=apiserver.Authorization.Mode=RBAC"`
        + " --extra-config=apiserver.GenericServerRunOptions.CorsAllowedOriginList='.*'"`
        + " --extra-config=controller-manager.ClusterSigningCertFile='/var/lib/localkube/certs/ca.crt'"`
        + " --extra-config=controller-manager.ClusterSigningKeyFile='/var/lib/localkube/certs/ca.key'"`
        + " --extra-config=apiserver.admission-control='LimitRanger,ServiceAccount,DefaultStorageClass,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota'"`
        + " --kubernetes-version=v${KUBERNETES_VERSION}"`
        + " --feature-gates='MountPropagation=false'"`
        + " --disk-size=${DISK_SIZE}"`
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

function IncreaseFsInotifyMaxUserInstances() {
    # Default value of 128 is not enough to perform “kubectl log -f” from pods, hence increased to 524288
    $cmd = "minikube ssh -- `"sudo sysctl -w fs.inotify.max_user_instances=524288`""
    Invoke-Expression -Command $cmd
}

function AddDevDomainsToEtcHosts([string[]]$hostnamesPrefixes) {
    $n = 6 # 7 hostnames in one line, others in next line. Windows can't read more than 9 hostnames in the same line.
    $hostnames = $hostnamesPrefixes | ForEach-Object {"$_.${DOMAIN}"} # for minikube ssh
    $hostnames1 = $hostnamesPrefixes[0..$n] | ForEach-Object {"$_.${DOMAIN}"}
    $hostnames2 = $hostnamesPrefixes[ - ($n + 1)..-1] | ForEach-Object {"$_.${DOMAIN}"}

    $cmd = "minikube ip"
    $minikubeIp = (Invoke-Expression -Command $cmd | Out-String).Trim()

    Write-Output "Minikube IP address: ${minikubeIp}"

    $cmd = "minikube ssh 'echo `"127.0.0.1 ${hostnames}`" | sudo tee -a /etc/hosts'"
    Invoke-Expression -Command $cmd

    $winHostsPath = "C:\Windows\system32\drivers\etc\hosts"

    Get-Content -Path $winHostsPath | Select-String -Pattern $DOMAIN -NotMatch | Out-File -FilePath $winHostsPath

    "${minikubeIp} ${hostnames1}" | Out-File $winHostsPath -Append
    "${minikubeIp} ${hostnames2}" | Out-File $winHostsPath -Append
}

CheckIfMinikubeIsInitialized
InitializeMinikubeConfig
StartMinikube
WaitForMinikubeToBeUp
AddDevDomainsToEtcHosts "apiserver", "console", "catalog", "instances", "dex", "docs", "lambdas-ui", "ui-api", "minio", "jaeger", "grafana", "configurations-generator", "gateway", "connector-service"
IncreaseFsInotifyMaxUserInstances
