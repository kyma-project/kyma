params (
    [string]$COMPONENT_NAME,
    [bool]$ENABLED
)

$CURRENT_DIR = Split-Path $MyInvocation.MyCommand.Path
$STATUS = $ENABLED ? "enabled" : "disabled"
$FILE_NAME = "components.env"
$FILE_PATH = "${CURRENT_DIR}\..\${FILE_NAME}"

# Create the file if it does not exists
if(![System.IO.File]::Exists($FILE_PATH)){
    Write-Output "Generating ${FILE_NAME} file"
    New-Item $FILE_PATH -Type File
}

# Remove previous entry in case the provided key exists
Get-Content -Path $FILE_PATH | Select-String -Pattern $COMPONENT_NAME -NotMatch | Out-File -FilePath $FILE_PATH

# Append the provided key and value to the file
"${COMPONENT_NAME}.enabled=${ENABLED}" | Out-File $FILE_PATH -Append

Write-Output "Component ${COMPONENT_NAME} is now ${STATUS}!"