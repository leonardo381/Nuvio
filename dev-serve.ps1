$ErrorActionPreference = "Stop"

$ListenAddress = "127.0.0.1"
$ListenPort = 8090
$Script:StartedProcessId = $null

function Import-EnvFile {
    param(
        [Parameter(Mandatory = $true)]
        [string]$FilePath
    )

    if (-not (Test-Path $FilePath)) {
        return
    }

    Get-Content $FilePath | ForEach-Object {
        $line = $_.Trim()

        if (-not [string]::IsNullOrWhiteSpace($line) -and -not $line.StartsWith("#")) {
            $separatorIndex = $line.IndexOf("=")
            if ($separatorIndex -ge 1) {
                $key = $line.Substring(0, $separatorIndex).Trim()
                $value = $line.Substring($separatorIndex + 1).Trim()

                if (
                    ($value.StartsWith('"') -and $value.EndsWith('"')) -or
                    ($value.StartsWith("'") -and $value.EndsWith("'"))
                ) {
                    $value = $value.Substring(1, $value.Length - 2)
                }

                [System.Environment]::SetEnvironmentVariable($key, $value, "Process")
            }
        }
    }
}

function Get-ProcessInfo {
    param([Parameter(Mandatory = $true)][int]$ProcessId)
    Get-CimInstance Win32_Process -Filter "ProcessId=$ProcessId" -ErrorAction SilentlyContinue
}

function Get-NormalizedCommandLine {
    param($Process)
    if (-not $Process -or -not $Process.CommandLine) {
        return ""
    }

    return ($Process.CommandLine -replace "\\", "/").ToLowerInvariant()
}

function Test-IsGoRunBaseProcess {
    param($Process)
    if (-not $Process -or $Process.Name -ine "go.exe") {
        return $false
    }

    $commandLine = Get-NormalizedCommandLine $Process
    return $commandLine.Contains(" run ./examples/base serve")
}

function Test-IsConfirmedNuvioCmsProcess {
    param($Process)
    if (-not $Process) {
        return $false
    }

    if (Test-IsGoRunBaseProcess $Process) {
        return $true
    }

    $commandLine = Get-NormalizedCommandLine $Process

    if ($Process.Name -ieq "base.exe" -and $commandLine.Contains("base.exe serve")) {
        $parent = Get-ProcessInfo -ProcessId $Process.ParentProcessId
        return (Test-IsConfirmedNuvioCmsProcess $parent)
    }

    if ($Process.Name -ieq "nuvio.exe" -and $commandLine.Contains(" serve")) {
        $executablePath = "$($Process.ExecutablePath)"
        return $executablePath.StartsWith($PSScriptRoot, [System.StringComparison]::OrdinalIgnoreCase)
    }

    return $false
}

function Show-ProcessDetails {
    param($Process)
    if (-not $Process) {
        return
    }

    Write-Host "PID: $($Process.ProcessId)"
    Write-Host "Name: $($Process.Name)"
    Write-Host "Executable: $($Process.ExecutablePath)"
    Write-Host "Command: $($Process.CommandLine)"
    Write-Host "Parent PID: $($Process.ParentProcessId)"

    $parent = Get-ProcessInfo -ProcessId $Process.ParentProcessId
    if ($parent) {
        Write-Host "Parent name: $($parent.Name)"
        Write-Host "Parent command: $($parent.CommandLine)"
    }
}

function Get-NuvioCmsStopRootProcessId {
    param([Parameter(Mandatory = $true)][int]$ProcessId)

    $current = Get-ProcessInfo -ProcessId $ProcessId
    if (-not $current) {
        return $null
    }

    $rootProcessId = $current.ProcessId

    while ($current) {
        if (Test-IsGoRunBaseProcess $current) {
            $rootProcessId = $current.ProcessId
        }

        $parent = Get-ProcessInfo -ProcessId $current.ParentProcessId
        if (-not (Test-IsConfirmedNuvioCmsProcess $parent)) {
            break
        }

        $current = $parent
    }

    return $rootProcessId
}

function Stop-ProcessTree {
    param(
        [Parameter(Mandatory = $true)][int]$ProcessId,
        [Parameter(Mandatory = $true)][string]$Reason
    )

    $process = Get-ProcessInfo -ProcessId $ProcessId
    if (-not $process) {
        return
    }

    Write-Host "Stopping NuvioCMS process tree PID $ProcessId ($Reason)."
    taskkill /PID $ProcessId /T /F | Out-Host
}

function Test-DevPortAvailable {
    $listener = Get-NetTCPConnection -LocalAddress $ListenAddress -LocalPort $ListenPort -State Listen -ErrorAction SilentlyContinue | Select-Object -First 1
    if (-not $listener) {
        return $true
    }

    $process = Get-ProcessInfo -ProcessId $listener.OwningProcess
    if (Test-IsConfirmedNuvioCmsProcess $process) {
        Write-Host "NuvioCMS already running on $ListenAddress`:$ListenPort, PID $($process.ProcessId)."
        Write-Host "Use the existing server or stop it with .\stop-dev.ps1 before starting a fresh one."
        Show-ProcessDetails $process
        exit 0
    }

    Write-Warning "Port $ListenAddress`:$ListenPort is occupied by a process that is not confirmed as local NuvioCMS."
    Show-ProcessDetails $process
    Write-Warning "Not killing this process. Free the port manually or choose the correct service."
    exit 1
}

Push-Location $PSScriptRoot
try {
    Import-EnvFile ".env"
    Import-EnvFile "ui/.env"

    Test-DevPortAvailable | Out-Null

    Push-Location "ui"
    try {
        npm run build
        if ($LASTEXITCODE -ne 0) {
            exit $LASTEXITCODE
        }
    } finally {
        Pop-Location
    }

    $startInfo = [System.Diagnostics.ProcessStartInfo]::new()
    $startInfo.FileName = "go"
    $startInfo.Arguments = "run ./examples/base serve"
    $startInfo.WorkingDirectory = $PSScriptRoot
    $startInfo.UseShellExecute = $false

    $process = [System.Diagnostics.Process]::Start($startInfo)
    $Script:StartedProcessId = $process.Id
    Write-Host "Started NuvioCMS dev server with go PID $($process.Id)."

    try {
        while (-not $process.WaitForExit(500)) {
            # Keep the wrapper alive so finally can clean up on interruption.
        }
        exit $process.ExitCode
    } finally {
        if ($process -and -not $process.HasExited) {
            Stop-ProcessTree -ProcessId $process.Id -Reason "dev-serve.ps1 exit"
        }
    }
} finally {
    if ($Script:StartedProcessId) {
        $startedProcess = Get-ProcessInfo -ProcessId $Script:StartedProcessId
        if ($startedProcess) {
            Stop-ProcessTree -ProcessId $Script:StartedProcessId -Reason "final cleanup"
        }
    }
    Pop-Location
}