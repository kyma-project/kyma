param (
    [string]$URL = "",
    [string]$OUTPUT = "",
    [string]$VERSION = "0.0.1"
)

$CURRENT_DIR = Split-Path $MyInvocation.MyCommand.Path
$CRTPL_PATH = "${CURRENT_DIR}\..\resources\installer-cr.yaml.tpl"

Copy-Item -Path $CRTPL_PATH -Destination $OUTPUT

(Get-Content $OUTPUT).replace("__VERSION__", $VERSION) | Set-Content $OUTPUT
(Get-Content $OUTPUT).replace("__URL__", $URL) | Set-Content $OUTPUT
