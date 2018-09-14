$CURRENT_DIR = Split-Path $MyInvocation.MyCommand.Path
$VERSIONS_FILE_URL = "https://kymainstaller.blob.core.windows.net/kyma-versions/latest.env"
$VERSIONS_FILE_PATH = "${CURRENT_DIR}\..\versions.env"

if(![System.IO.File]::Exists($VERSIONS_FILE_PATH) -or ((Get-Content $VERSIONS_FILE_PATH).Length -eq 0)) {
    Write-Output "Downloading versions.env file."
    Invoke-WebRequest ${VERSIONS_FILE_URL} -OutFile ${VERSIONS_FILE_PATH}
} else {
    Write-Output "File ${VERSIONS_FILE_PATH} exists, reusing."
}
