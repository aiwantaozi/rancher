<#
	output.ps1 could output the execution commands, and then
	pass the content to `Invoke-Expression` via pipe character
 #>

$ErrorActionPreference = 'Stop'
$WarningPreference = 'SilentlyContinue'
$VerbosePreference = 'SilentlyContinue'
$DebugPreference = 'SilentlyContinue'
$InformationPreference = 'SilentlyContinue'

try
{
    New-Item -Force -Type Directory -Path @(
        "c:\host\opt"
        "c:\host\opt\cni"
        "c:\host\opt\cni\bin"
        "c:\host\etc"
        "c:\host\etc\rancher"
        "c:\host\etc\kubernetes"
        "c:\host\etc\cni"
        "c:\host\etc\cni\net.d"
        "c:\host\etc\nginx"
        "c:\host\etc\nginx\logs"
        "c:\host\etc\nginx\temp"
        "c:\host\etc\nginx\conf"
        "c:\host\etc\kube-flannel"
        "c:\host\var"
        "c:\host\var\run"
        "c:\host\var\log"
        "c:\host\var\log\pods"
        "c:\host\var\log\containers"
        "c:\host\var\lib"
        "c:\host\var\lib\cni"
        "c:\host\var\lib\rancher"
        "c:\host\var\lib\kubelet"
        "c:\host\var\lib\kubelet\volumeplugins"
        "c:\host\run"
    ) | Out-Null
} catch { }

try
{
    Copy-Item -Force -Destination "c:\host\etc\rancher" -Path @(
        "c:\etc\rancher\cleanup.ps1"
        "c:\Windows\wins.exe"
    )
} catch { }


$entryArgs = @("run", "-d", "--restart=unless-stopped", "-v", "$($env:CUSTOM_DOCKER_NAMED_PIPE)://./pipe/docker_engine", "-v", "//./pipe/rancher_wins://./pipe/rancher_wins", "-v", "c:/etc/kubernetes:c:/etc/kubernetes", $env:AGENT_IMAGE, "execute") + $args
Out-File -Encoding ascii -FilePath "c:\host\etc\rancher\bootstrap.ps1" -InputObject @"
function Log-Warn
{
    Write-Host -NoNewline -ForegroundColor DarkYellow "WARN: "
    Write-Host -ForegroundColor Gray ("{0,-44}" -f (`$args -join " "))
}

function Log-Fatal
{
    Write-Host -NoNewline -ForegroundColor DarkRed "FATA: "
    Write-Host -ForegroundColor Gray ("{0,-44}" -f (`$args -join " "))

    exit 1
}

function Is-Administrator
{
    `$p = New-Object System.Security.Principal.WindowsPrincipal([System.Security.Principal.WindowsIdentity]::GetCurrent())
    return `$p.IsInRole([System.Security.Principal.WindowsBuiltInRole]::Administrator)
}

# TODO add cpu ram verification

# check identity
if (-not (Is-Administrator))
{
    Log-Fatal "You need elevated Administrator privileges in order to run this script, start Windows PowerShell by using the Run as Administrator option"
}

# check docker npipe
`$dockerNPipe = Get-ChildItem //./pipe/ -ErrorAction Ignore | ? Name -eq "docker_engine"
if (-not `$dockerNPipe)
{
    Log-Warn "Default docker named pipe (//./pipe/docker_engine) is not found"
    Log-Warn "Please indicate the custom docker named pipe environmental variable if docker errors occur"
    Log-Warn "example: docker run -e CUSTOM_DOCKER_NAMED_PIPE=//./pipe/custom_docker_named_pipe ..."
}

# check docker release
try
{
    `$dockerPlatform = docker.exe version -f "{{.Server.Platform.Name}}"
    if (-not (`$dockerPlatform -like '*Enterprise*'))
    {
        Log-Fatal "Only support with Docker EE"
    }
}
catch
{
    Log-Fatal "Could not found Docker service: `$(`$_.Exception.Message)"
}

# check system locale
`$sysLocale = Get-WinSystemLocale | Select-Object -ExpandProperty "IetfLanguageTag"
if (-not `$sysLocale.StartsWith('en-'))
{
    Log-Fatal "Only support with English System Locale"
}

# check network count
`$vNetAdapters = Get-HnsNetwork | Select-Object -ExpandProperty "Subnets" | Select-Object -ExpandProperty "GatewayAddress"
`$allNetAdapters = Get-WmiObject -Class Win32_NetworkAdapterConfiguration -Filter "IPEnabled=True" | Sort-Object Index | ForEach-Object { `$_.IPAddress[0] } | Where-Object { -not (`$vNetAdapters -contains `$_) }
`$networkCount = `$allNetAdapters | Measure-Object | Select-Object -ExpandProperty "Count"
if (`$networkCount -gt 1)
{
    Log-Warn "More than 1 network interfaces are found: `$(`$allNetAdapters -join ", ")"
    Log-Warn "Please indicate --internal-address when adding failed"
}

# check msiscsi servcie running
`$svcMsiscsi = Get-Service -Name "msiscsi" -ErrorAction Ignore
if (`$svcMsiscsi -and (`$svcMsiscsi.Status -ne "Running"))
{
    Set-Service -Name "msiscsi" -StartupType Automatic -WarningAction Ignore
    Start-Service -Name "msiscsi" -ErrorAction Ignore -WarningAction Ignore
    if (-not `$?) {
        Log-Warn "Failed to start msiscsi service, you may not be able to use the iSCSI flexvolume properly"
    }
}

# repair Get-GcePdName method
# this's a stopgap, we could drop this after https://github.com/kubernetes/kubernetes/issues/74674 fixed
# related: rke-tools, hyperkube
`$getGcePodNameCommand = Get-Command -Name "Get-GcePdName" -ErrorAction Ignore
if (-not `$getGcePodNameCommand) {
    `$profilePath = "`$PsHome\profile.ps1"
    if (-not (Test-Path `$profilePath)) {
        New-Item -Path `$profilePath -Type File -ErrorAction Ignore | Out-Null
    }
    `$appendProfile = @'
Unblock-File -Path DLLPATH -ErrorAction Ignore
Import-Module -Name DLLPATH -ErrorAction Ignore
'@
    Add-Content -Path `$profilePath -Value `$appendProfile.replace('DLLPATH', "c:\etc\kubernetes\GetGcePdName.dll") -ErrorAction Ignore
}

## register wins
## TODO need TLS
#`$winsRegister = Execute-Binary -FilePath "c:\etc\rancher\wins.exe" -ArgumentList "srv register"
#if (-not `$winsRegister.Success)
#{
#    Log-Fatal "Failed to register rancher-wins service:" `$winsRegister.StdErr
#}
#
## start wins
#Start-Service -Name "rancher-wins" -ErrorAction Ignore
#if (-not `$?)
#{
#    Log-Fatal "Failed to start rancher-wins service"
#}

# test
Start-Process ``
    -WindowStyle Hidden ``
    -FilePath "c:\etc\rancher\wins.exe" ``
    -ArgumentList "srv serve --debug"

# run agent
Start-Process -NoNewWindow -Wait ``
    -FilePath "docker.exe" ``
    -ArgumentList "$($entryArgs -join " ")"

# remove script
Remove-Item -Force -Path "c:\etc\rancher\bootstrap.ps1" -ErrorAction Ignore
"@

Write-Output -InputObject "c:\etc\rancher\bootstrap.ps1"
