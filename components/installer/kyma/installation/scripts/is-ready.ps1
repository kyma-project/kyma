param (
    [string]$ns,
    [string]$label,
    [string]$value
)

$cmdPods = "kubectl.exe get pods -n ${ns} -l ${label}=${value} -o jsonpath='{.items[*].metadata.name}'"

# Checking if POD is already deployed
while($true) {
    $pods = (Invoke-Expression -Command $cmdPods | Out-String).Trim()
    if ($pods -ne "") {
      Write-Output "${value} is deployed..."
      break
    }
    else {
      Write-Output "${value} is not deployed - waiting 5s..."
      Start-Sleep 5
    }
}

# Checking if POD is ready to operate
$pods = (Invoke-Expression -Command $cmdPods | Out-String).Trim().Split("`n")
for($i = 0; $i -lt $pods.Length; $i += 1) {
    $pod = $pods[$i]
    while($true) {
        $cmdPodStatus = "kubectl get pod ${pod} -n ${ns} -o jsonpath='{.status.containerStatuses[0].ready}'"
        $podStatus = (Invoke-Expression -Command $cmdPodStatus | Out-String).Trim()
        if ($podStatus -eq "true") {
            Write-Output "${pod} is running"
            break
        }
        else {
            Write-Output "${pod} is not running - waiting 5s..."
            Start-Sleep 5
        }
    }
}

# Checking only if kube-dns is checked
if ("${value}" -eq "kube-dns") {
    while($true) {
        $ipCmd = "kubectl.exe get ep ${value} -n ${ns} -o jsonpath='{.subsets[0].addresses[0].ip}'"
        $ip = (Invoke-Expression $ipCmd | Out-String).Trim()
        if($ip -ne "") {
            Write-Output "kubedns endpoint IP assigned - ${ip}"
            break
        }
        else {
            Write-Output "kubedns endpoint IP is not assigned yet - waiting 5s..."
            Start-Sleep 5
        }
    }
}