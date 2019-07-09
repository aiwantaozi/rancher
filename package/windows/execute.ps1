<#
	run.ps1 executes the agent
 #>

$ErrorActionPreference = 'Stop'
$WarningPreference = 'SilentlyContinue'
$VerbosePreference = 'SilentlyContinue'
$DebugPreference = 'SilentlyContinue'
$InformationPreference = 'SilentlyContinue'

function Log-Info
{
    Write-Host -NoNewline -ForegroundColor Blue "INFO: "
    Write-Host -ForegroundColor Gray ("{0,-44}" -f ($args -join " "))
}

function Log-Warn
{
    Write-Host -NoNewline -ForegroundColor DarkYellow "WARN: "
    Write-Host -ForegroundColor Gray ("{0,-44}" -f ($args -join " "))
}

function Log-Error
{
    Write-Host -NoNewline -ForegroundColor DarkRed "ERRO: "
    Write-Host -ForegroundColor Gray ("{0,-44}" -f ($args -join " "))
}

function Log-Fatal
{
    Write-Host -NoNewline -ForegroundColor DarkRed "FATA: "
    Write-Host -ForegroundColor Gray ("{0,-44}" -f ($args -join " "))

    exit 1
}

function Set-Env
{
    param(
        [parameter(Mandatory = $true)] [string]$Key,
        [parameter(Mandatory = $false)] [string]$Value = ""
    )

    try {
        [Environment]::SetEnvironmentVariable($Key, $Value, [EnvironmentVariableTarget]::Process)
    } catch {
        Log-Error "Could not set $Key = $Value in Process target: $($_.Exception.Message)"
    }

    try {
        [Environment]::SetEnvironmentVariable($Key, $Value, [EnvironmentVariableTarget]::Machine)
    } catch {
        Log-Error "Could not set $Key = $Value in Machine target: $($_.Exception.Message)"
    }
}

function Get-Env
{
    param(
        [parameter(Mandatory = $true)] [string]$Key
    )

    try {
        $val = [Environment]::GetEnvironmentVariable($Key, [EnvironmentVariableTarget]::Process)
        if ($val) {
            return $val
        }
    } catch {
        Log-Error "Could not get $Key in Process target: $($_.Exception.Message)"
    }

    try {
        $val = [Environment]::GetEnvironmentVariable($Key, [EnvironmentVariableTarget]::Machine)
        if ($val) {
            return $val
        }
    } catch {
        Log-Error "Could not get $Key in Machine target: $($_.Exception.Message)"
    }

    return ""
}

function ConvertTo-JsonObj
{
    param (
        [parameter(Mandatory = $false, ValueFromPipeline = $true)] [string]$JSON
    )

    $ret = @{}
    try {
        $jsonConfig = $JSON | ConvertFrom-Json -ErrorAction Ignore -WarningAction Ignore
        $jsonConfig.PSObject.Properties | ForEach-Object {
            $item = $_
            $ret[$item.Name] = $item.Value
        }
    } catch {}

    return $ret
}

function Get-Address
{
    param(
        [parameter(Mandatory = $false)] [string]$Addr
    )

    if (-not $Addr) {
        return ""
    }

    # TODO If given address is a network interface on the system, retrieve configured IP on that interface (only the first configured IP is taken)
#    try
#    {
#        $na = Get-NetAdapter | ? Name -eq $Addr
#        if ($na)
#        {
#            return (Get-NetIPAddress -InterfaceIndex $na.ifIndex -AddressFamily IPv4).IPAddress
#        }
#    }
#    catch { }

    # Repair the container route for `169.254.169.254` before cloud provider query
    $actualGateway = route print 0.0.0.0 | Where-Object {$_ -match '0\.0\.0\.0.*[a-z]'} | Select-Object -First 1 | ForEach-Object {($_ -replace '0\.0\.0\.0|[a-z]|\s+',' ').Trim() -split ' '} | Select-Object -First 1
    $expectedGateway = route print 169.254.169.254 | Where-Object {$_ -match '169\.254\.169\.254'} | Select-Object -First 1 | ForEach-Object {($_ -replace '169\.254\.169\.254|255\.255\.255\.255|[a-z]|\s+',' ').Trim() -split ' '} | Select-Object -First 1
    if ($actualGateway -ne $expectedGateway) {
        route add 169.254.169.254 MASK 255.255.255.255 $actualGateway METRIC 1 | Out-Null
    }

    # Loop through cloud provider options to get IP from metadata, if not found return given value
    switch ($Addr)
    {
        "awslocal" {
            return $(curl.exe -s "http://169.254.169.254/latest/meta-data/local-ipv4")
        }
        "awspublic" {
            return $(curl.exe -s "http://169.254.169.254/latest/meta-data/public-ipv4")
        }
        "doprivate" {
            return $(curl.exe -s "http://169.254.169.254/metadata/v1/interfaces/private/0/ipv4/address")
        }
        "dopublic" {
            return $(curl.exe -s "http://169.254.169.254/metadata/v1/interfaces/public/0/ipv4/address")
        }
        "azprivate" {
            return $(curl -s -H "Metadata:true" "http://169.254.169.254/metadata/instance/network/interface/0/ipv4/ipAddress/0/privateIpAddress?api-version=2017-08-01&format=text")
        }
        "azpublic" {
            return $(curl -s -H "Metadata:true" "http://169.254.169.254/metadata/instance/network/interface/0/ipv4/ipAddress/0/publicIpAddress?api-version=2017-08-01&format=text")
        }
        "gceinternal" {
            return $(curl -s -H "Metadata-Flavor: Google" "http://metadata.google.internal/computeMetadata/v1/instance/network-interfaces/0/ip?alt=json")
        }
        "gceexternal" {
            return $(curl -s -H "Metadata-Flavor: Google" "http://metadata.google.internal/computeMetadata/v1/instance/network-interfaces/0/access-configs/0/external-ip?alt=json")
        }
        "packetlocal" {
            return $(curl -s "https://metadata.packet.net/2009-04-04/meta-data/local-ipv4")
        }
        "packetpublic" {
            return $(curl -s "https://metadata.packet.net/2009-04-04/meta-data/public-ipv4")
        }
        "ipify" {
            return $(curl -s "https://api.ipify.org")
        }
    }

    return $Addr
}

# required envs
Set-Env -Key "DOCKER_HOST" -Value "npipe:////./pipe/docker_engine"
Set-Env -Key "CATTLE_ROLE" -Value "worker"

# clean up
$CLUSTER_CLEANUP = Get-Env -Key "CLUSTER_CLEANUP"
if ($CLUSTER_CLEANUP -eq "true")
{
    Start-Process -NoNewWindow -Wait -FilePath "c:\etc\rancher\agent.exe"
    exit 0
}

# init parameters
$CATTLE_SERVER = Get-Env -Key "CATTLE_SERVER"
$CATTLE_TOKEN = Get-Env -Key "CATTLE_TOKEN"
$CATTLE_NODE_NAME = Get-Env -Key "CATTLE_NODE_NAME"
$CATTLE_ADDRESS = Get-Env -Key "CATTLE_ADDRESS"
$CATTLE_INTERNAL_ADDRESS = Get-Env -Key "CATTLE_INTERNAL_ADDRESS"
$CATTLE_CA_CHECKSUM = Get-Env -Key "CATTLE_CA_CHECKSUM"
$CATTLE_NODE_LABEL = @()

# parse parameters
$vals = $null
for ($i = $args.Length; $i -ge 0; $i--)
{
    $arg = $args[$i]
    switch -regex ($arg)
    {
        '(-d|--debug)' {
            Set-Env -Key "CATTLE_DEBUG" -Value "true"
            $vals = $null
        }
        '(-s|--server)' {
            $CATTLE_SERVER = ($vals | Select-Object -Last 1)
            $vals = $null
        }
        '(-t|--token)' {
            $CATTLE_TOKEN = ($vals | Select-Object -Last 1)
            $vals = $null
        }
        '(-c|--ca-checksum)' {
            $CATTLE_CA_CHECKSUM = ($vals | Select-Object -Last 1)
            $vals = $null
        }
        '(-all|--all-roles)' {
            $vals = $null
        }
        '(-e|--etcd)' {
            $vals = $null
        }
        '(-w|--worker)' {
            $vals = $null
        }
        '(-p|--controlplane)' {
            $vals = $null
        }
        '(-r|--node-name)' {
            $CATTLE_NODE_NAME = ($vals | Select-Object -Last 1)
            $vals = $null
        }
        '(-n|--no-register)' {
            Set-Env -Key "CATTLE_AGENT_CONNECT" -Value "true"
            $vals = $null
        }
        '(-a|--address)' {
            $CATTLE_ADDRESS = ($vals | Select-Object -Last 1)
            $vals = $null
        }
        '(-i|--internal-address)' {
            $CATTLE_INTERNAL_ADDRESS = ($vals | Select-Object -Last 1)
            $vals = $null
        }
        '(-l|--label)' {
            $CATTLE_NODE_LABEL = $vals
            $vals = $null
        }
        '(-o|--only-write-certs)' {
            Set-Env -Key "CATTLE_WRITE_CERT_ONLY" -Value "true"
            $vals = $null
        }
        default {
            if ($vals) {
                $vals += @($arg)
            } else {
                $vals = @($arg)
            }
        }
    }
}

# check docker npipe
$CATTLE_CLUSTER = Get-Env -Key "CATTLE_CLUSTER"
if ($CATTLE_CLUSTER -ne "true")
{
    $dockerNPipe = Get-ChildItem //./pipe/ -ErrorAction Ignore | ? Name -eq "docker_engine"
    if (-not $dockerNPipe) {
        Log-Warn "Default docker named pipe is not found"
        Log-Warn "Please bind mount in the docker named pipe to //./pipe/docker_engine if docker errors occur"
        Log-Warn "example: docker run -v //./pipe/custom_docker_named_pipe://./pipe/docker_engine ..."
    }
}

# get default network metadata when nodeName or address is blank
if ((-not $CATTLE_NODE_NAME) -or (-not $CATTLE_ADDRESS))
{
    $getAdapterJson = wins.exe cli net get-adapter
    if (-not $?) {
        Log-Fatal "Could not get host network metadata"
    }
    $defaultNetwork = $getAdapterJson | ConvertFrom-Json
    if (-not $defaultNetwork) {
        Log-Fatal "Could not get host network metadata"
    }

    if (-not $CATTLE_NODE_NAME) {
        $CATTLE_NODE_NAME = $defaultNetwork.HostName
        $CATTLE_NODE_NAME = $CATTLE_NODE_NAME.ToLower()
    }

    if (-not $CATTLE_ADDRESS) {
        $CATTLE_ADDRESS = $defaultNetwork.AddressCIDR -replace "/32",""
    }
}

# get address
$CATTLE_ADDRESS = Get-Address -Addr $CATTLE_ADDRESS
$CATTLE_INTERNAL_ADDRESS = Get-Address -Addr $CATTLE_INTERNAL_ADDRESS

# check token and address
$CATTLE_K8S_MANAGED = Get-Env -Key "CATTLE_K8S_MANAGED"
if ($CATTLE_K8S_MANAGED -ne "true")
{
    if (-not $CATTLE_TOKEN) {
        Log-Fatal "--token is a required option"
    }
    if (-not $CATTLE_ADDRESS) {
        Log-Fatal "--address is a required option"
    }
}

# check rancher server address
if (-not $CATTLE_SERVER)
{
    Log-Fatal "--server is a required option"
}

# check rancher server
try
{
    curl.exe --insecure -s -fL "$CATTLE_SERVER/ping" | Out-Null
    if ($?) {
        Log-Info "$CATTLE_SERVER is accessible"
    } else {
        Log-Fatal "$CATTLE_SERVER is not accessible"
    }
}
catch
{
    Log-Fatal "$CATTLE_SERVER is not accessible: $($_.Exception.Message)"
}

# download cattle server CA
if ($CATTLE_CA_CHECKSUM)
{
    $sslCertDir = Get-Env -Key "SSL_CERT_DIR"
    $server = $CATTLE_SERVER
    $caChecksum = $CATTLE_CA_CHECKSUM
    $temp = New-TemporaryFile
    $cacerts = $null
    try {
        $cacerts = $(curl.exe --insecure -s -fL "$server/v3/settings/cacerts" | ConvertTo-JsonObj).value
    } catch {}
    if (-not $cacerts) {
        Log-Fatal "Could not get cattle server CA from $server"
    }

    $cacerts + "`n" | Out-File -NoNewline -Encoding ascii -FilePath $temp.FullName
    $tempHasher = Get-FileHash -LiteralPath $temp.FullName -Algorithm SHA256
    if ($tempHasher.Hash.ToLower() -ne $caChecksum.ToLower()) {
        $temp.Delete()
        Log-Fatal "Actual cattle server CA checksum is $($tempHasher.Hash.ToLower()), $server/v3/settings/cacerts does not match $($caChecksum.ToLower())"
    }
    Remove-Item -Force -Recurse -Path "$sslCertDir\serverca" -ErrorAction Ignore
    New-Item -Force -Type Directory -Path $sslCertDir -ErrorAction Ignore | Out-Null
    $temp.MoveTo("$sslCertDir\serverca")

    # import the self-signed certificate
    certoc.exe -addstore root "$sslCertDir\serverca" | Out-Null
    if (-not $?) {
        Log-Error "Failed to import rancher server certificates to Root"
    }

    $CATTLE_SERVER_HOSTNAME = ([System.Uri]"$server").Host
    $CATTLE_SERVER_HOSTNAME_WITH_PORT = ([System.Uri]"$server").Authority

    $dockerCertsPath = "c:\etc\docker\certs.d\$CATTLE_SERVER_HOSTNAME_WITH_PORT"
    New-Item -Force -Type Directory -Path $dockerCertsPath -ErrorAction Ignore | Out-Null
    Copy-Item -Force -Path "$sslCertDir\serverca" -Destination "$dockerCertsPath\ca.crt" -ErrorAction Ignore
}

# add labels
Get-ItemProperty 'HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\' -ErrorAction Ignore | ForEach-Object {
    $versionTag = "$($windowsCurrentVersion.CurrentMajorVersionNumber).$($windowsCurrentVersion.CurrentMinorVersionNumber).$($windowsCurrentVersion.CurrentBuildNumber).$($windowsCurrentVersion.UBR)"
    $CATTLE_NODE_LABEL += @("rke.cattle.io/windows-version=$versionTag")
    $CATTLE_NODE_LABEL += @("rke.cattle.io/windows-release-id=$($windowsCurrentVersion.ReleaseId)")
    $CATTLE_NODE_LABEL += @("rke.cattle.io/windows-major-version=$($windowsCurrentVersion.CurrentMajorVersionNumber)")
    $CATTLE_NODE_LABEL += @("rke.cattle.io/windows-minor-version=$($windowsCurrentVersion.CurrentMinorVersionNumber)")
    $CATTLE_NODE_LABEL += @("rke.cattle.io/windows-kernel-version=$($windowsCurrentVersion.BuildLabEx)")
    $CATTLE_NODE_LABEL += @("rke.cattle.io/windows-build=$($windowsCurrentVersion.CurrentBuild)")
}

# set environment variables
Set-Env -Key "CATTLE_SERVER" -Value $CATTLE_SERVER
Set-Env -Key "CATTLE_TOKEN" -Value $CATTLE_TOKEN
Set-Env -Key "CATTLE_ADDRESS" -Val $CATTLE_ADDRESS
Set-Env -Key "CATTLE_INTERNAL_ADDRESS" -Val $CATTLE_INTERNAL_ADDRESS
Set-Env -Key "CATTLE_NODE_NAME" -Value $CATTLE_NODE_NAME
Set-Env -Key "CATTLE_NODE_LABEL" -Value $($CATTLE_NODE_LABEL -join ",")

Start-Process -NoNewWindow -Wait -FilePath "c:\etc\rancher\agent.exe"
