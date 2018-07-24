param(
    [switch]$VERBOSE
)

$DELAY = 10
if ($VERBOSE) {
    $DELAY = 5
}

Write-Output "Checking state of kyma installation...hold on"

while($true) {
    $statusCmd = "kubectl.exe get installation/kyma-installation -o jsonpath='{.status.state}'"
    $status = (Invoke-Expression -Command $statusCmd | Out-String).Trim()
  
    $descCmd = "kubectl.exe get installation/kyma-installation -o jsonpath='{.status.description}'"
    $desc = (Invoke-Expression -Command $descCmd | Out-String).Trim()
    if ($status -eq "Installed") {
        Write-Output "kyma is installed..."
        break
    } elseif ($status -eq "Error") {
        Write-Output "kyma installation error... ${desc}"
        Write-Output "----------"
        $cmd = "kubectl.exe -n ${INSTALLER_NS} get pods -o jsonpath='{.items[*].metadata.name}'"
        $podName = (Invoke-Expression -Command $cmd | Out-String).ToString().Trim()
        
        $cmd = "kubectl.exe logs ${podName} -n kyma-installer"
        (Invoke-Expression -Command $cmd | Out-String).Trim() | Write-Output
        
        exit 1
    } else {
        Write-Output "Status: ${status}, description: ${desc}"
        if ($VERBOSE) {
            $cmd = "kubectl.exe get installation/kyma-installation -o yaml"
            (Invoke-Expression -Command $cmd | Out-String).Trim() | Write-Output
            
            Write-Output "----------"
            
            $cmd = "kubectl.exe get po --all-namespaces"
            (Invoke-Expression -Command $cmd | Out-String).Trim() | Write-Output
        }
        Start-Sleep $DELAY
    }
}