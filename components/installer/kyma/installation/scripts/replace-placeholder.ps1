param (
    [string]$PATH,
    [string]$PLACEHOLDER,
    [string]$VALUE
)

(Get-Content $PATH).replace("${PLACEHOLDER}", ${VALUE}) | Set-Content $PATH