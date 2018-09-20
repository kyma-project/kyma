param (
    [string]$COMPONENT,
    [switch]$DISABLED
)

$CURRENT_DIR = Split-Path $MyInvocation.MyCommand.Path
$STATUS = if($DISABLED) { "disabled" } else { "enabled" }
$STATE = if($DISABLED) { "false" } else { "true" }
$FILE_NAME = "components.env"
$FILE_PATH = "${CURRENT_DIR}\..\${FILE_NAME}"

# Create the file if it does not exists
if(![System.IO.File]::Exists($FILE_PATH)){
    Write-Output "Generating ${FILE_NAME} file"
    "" | Out-File -FilePath $FILE_PATH -Encoding utf8
}

# Remove previous entry in case the provided key exists
(Get-Content -Path $FILE_PATH | Select-String -Pattern $COMPONENT -NotMatch -Encoding utf8 | Out-String).Trim() | Out-File -FilePath $FILE_PATH -Encoding utf8

# Append the provided key and value to the file
"${COMPONENT}.enabled=${STATE}" | Out-File $FILE_PATH -Append -Encoding utf8

Write-Output "Component ${COMPONENT} is now ${STATUS}!"