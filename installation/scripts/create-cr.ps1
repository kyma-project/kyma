param (
    [string]$URL = "",
    [string]$OUTPUT = "",
    [string]$VERSION = "0.0.1",
    [string]$CRTPL_PATH = ""
)

$CURRENT_DIR = Split-Path $MyInvocation.MyCommand.Path

if ($CRTPL_PATH -eq "") {
    $CRTPL_PATH = "${CURRENT_DIR}\..\resources\installer-cr.yaml.tpl"
}

Copy-Item -Path $CRTPL_PATH -Destination $OUTPUT

(Get-Content $OUTPUT).replace("__VERSION__", $VERSION) | Set-Content $OUTPUT
(Get-Content $OUTPUT).replace("__URL__", $URL) | Set-Content $OUTPUT
