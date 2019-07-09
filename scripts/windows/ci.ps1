$ErrorActionPreference = 'Stop'

Invoke-Expression -Command "$PSScriptRoot\build.ps1"
Invoke-Expression -Command "$PSScriptRoot\test.ps1"
Invoke-Expression -Command "$PSScriptRoot\package.ps1"
