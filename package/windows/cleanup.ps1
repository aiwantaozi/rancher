<#
    clean all windows settings
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

function Is-Administrator
{
    $p = New-Object System.Security.Principal.WindowsPrincipal([System.Security.Principal.WindowsIdentity]::GetCurrent())
    return $p.IsInRole([System.Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Execute-Binary
{
    param (
        [parameter(Mandatory = $true)] [string]$FilePath,
        [parameter(Mandatory = $false)] [string[]]$ArgumentList
    )

    $stdout = New-TemporaryFile
    $stderr = New-TemporaryFile
    $stdoutContent = ""
    $stderrContent = ""
    try {
        if ($ArgumentList) {
            Start-Process -NoNewWindow -Wait `
                -FilePath $FilePath `
                -ArgumentList $ArgumentList `
                -RedirectStandardOutput $stdout.FullName `
                -RedirectStandardError $stderr.FullName `
                -ErrorAction Ignore
        } else {
            Start-Process -NoNewWindow -Wait `
                -FilePath $FilePath `
                -RedirectStandardOutput $stdout.FullName `
                -RedirectStandardError $stderr.FullName `
                -ErrorAction Ignore
        }

        $stdoutContent = (Get-Content $stdout.FullName)
        $stderrContent = (Get-Content $stderr.FullName)
    } catch {
        $stderrContent = $_.Exception.Message
    }

    $stdout.Delete()
    $stderr.Delete()

    return @{} | Add-Member -NotePropertyMembers @{
        StdOut = $stdoutContent
        StdErr = $stderrContent
        Success = [string]::IsNullOrEmpty($stderrContent)
    } -PassThru
}

function Get-VmComputeNativeMethods()
{
    $signature = @'
[DllImport("vmcompute.dll")]
public static extern void HNSCall([MarshalAs(UnmanagedType.LPWStr)] string method, [MarshalAs(UnmanagedType.LPWStr)] string path, [MarshalAs(UnmanagedType.LPWStr)] string request, [MarshalAs(UnmanagedType.LPWStr)] out string response);
'@

    # Compile into runtime type
    try {
        Add-Type -MemberDefinition $signature -Namespace VmCompute.PrivatePInvoke -Name NativeMethods -PassThru -ErrorAction Ignore
    } catch {}
}

function Invoke-HNSRequest
{
    param
    (
        [ValidateSet('GET', 'POST', 'DELETE')]
        [parameter(Mandatory = $true)] [string] $Method,
        [ValidateSet('networks', 'endpoints', 'activities', 'policylists', 'endpointstats', 'plugins')]
        [parameter(Mandatory = $true)] [string] $Type,
        [parameter(Mandatory = $false)] [string] $Action,
        [parameter(Mandatory = $false)] [string] $Data = "",
        [parameter(Mandatory = $false)] [Guid] $Id = [Guid]::Empty
    )

    $hnsPath = "/$Type"
    if ($id -ne [Guid]::Empty) {
        $hnsPath += "/$id"
    }
    if ($Action) {
        $hnsPath += "/$Action"
    }

    $response = ""
    $hnsApi = Get-VmComputeNativeMethods
    $hnsApi::HNSCall($Method, $hnsPath, "$Data", [ref]$response)

    $output = @()
    if ($response) {
        try {
            $output = ($response | ConvertFrom-Json)
            if ($output.Error) {
                Log-Error $output;
            } else {
                $output = $output.Output;
            }
        } catch {
            Log-Error $_.Exception.Message
        }
    }

    return $output;
}

Log-Info "Start cleanning ..."

# check identity
if (-not (Is-Administrator))
{
    Log-Fatal "You need elevated Administrator privileges in order to run this script, start Windows PowerShell by using the Run as Administrator option"
}

# remove rancher-wins service
Log-Info "Stopping rancher-wins service ..."
Get-Service -Name "rancher-wins" -ErrorAction Ignore | Where-Object {$_.Status -eq "Running"} | Stop-Service -Force -ErrorAction Ignore
Execute-Binary -FilePath "c:\etc\rancher\wins.exe" -ArgumentList @("srv", "unregister") | Where-Object {$_.Success -ne $true} | ForEach-Object {
    Log-Warn "Could not unregister rancher-wins service, $($_.StdErr)"
}

# stop kubernetes components processes
Get-Process -ErrorAction Ignore -Name @(
    "flanneld*"
    "kubelet*"
    "kube-proxy*"
    "nginx*"
    "wins"
    "RANCHER-WINS-*"
) | ForEach-Object {
    Log-Info "Stopping $( $_.Name ) process ..."
    $_ | Stop-Process -Force -ErrorAction Ignore
}

# clean up docker conatiner: docker rm -fv $(docker ps -qa)
Log-Info "Removing Docker containers ..."
Execute-Binary -FilePath "docker.exe" -ArgumentList @('ps', '-qa') | Select-Object -ExpandProperty "StdOut" | ForEach-Object {
    Start-Process -NoNewWindow -Wait -FilePath "docker.exe" -ArgumentList @('rm', '-fv', $_) -ErrorAction Ignore
}

# clean network interface
Log-Info "Removing network settings ..."
Get-HnsNetwork | Where-Object {@('cbr0', 'vxlan0') -contains $_.Name} | Remove-HnsNetwork
Invoke-HNSRequest -Type policylists -Method GET | ForEach-Object {
    Invoke-HNSRequest -Method DELETE -Type policylists -Id $_.Id
}

# clean firewall rules
Log-Info "Removing firewall rules ..."
Remove-NetFirewallRule -ErrorAction Ignore -Name @(
    'RANCHER-WINS-UDP-4789'
    'RANCHER-WINS-TCP-10250'
    'RANCHER-WINS-TCP-10256'
) | Out-Null

# backup
Log-Info "Backing up ..."
$date = (Get-Date).ToString('yyyyMMddHHmm')
Copy-Item -Recurse -Path "c:\etc\rancher" -Destination "c:\etc\rancher-bak-$date" -Exclude "connected" -Force -ErrorAction Ignore | Out-Null
Copy-Item -Recurse -Path "c:\etc\nginx" -Destination "c:\etc\nginx-bak-$date" -Force -ErrorAction Ignore | Out-Null
Copy-Item -Recurse -Path "c:\etc\cni" -Destination "c:\etc\cni-bak-$date" -Force -ErrorAction Ignore | Out-Null
Copy-Item -Recurse -Path "c:\etc\kubernetes" -Destination "c:\etc\kubernetes-bak-$date" -Force -ErrorAction Ignore | Out-Null

# clean up
Log-Info "Cleaning up ..."
Remove-Item -Recurse -Force -ErrorAction Ignore -Path @(
    "c:\opt\*"
    "c:\etc\rancher\*"
    "c:\etc\nginx\*"
    "c:\etc\cni\*"
    "c:\etc\kubernetes\*"
    "c:\var\run\*"
    "c:\var\log\*"
    "c:\var\lib\*"
    "c:\run\*"
) | Out-Null

# restart docker service
Log-Info "Restarting Docker service ..."
Restart-Service -Name "docker" -ErrorAction Ignore | Out-Null

Log-Info "Finished!!!"