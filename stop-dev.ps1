$ErrorActionPreference = "Stop"

$ListenAddress = "127.0.0.1"
$ListenPort = 8090

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
    param([Parameter(Mandatory = $true)][int]$ProcessId)

    $process = Get-ProcessInfo -ProcessId $ProcessId
    if (-not $process) {
        Write-Host "Process PID $ProcessId is no longer running."
        return
    }

    Write-Host "Stopping confirmed NuvioCMS process tree PID $ProcessId."
    taskkill /PID $ProcessId /T /F | Out-Host
}

Push-Location $PSScriptRoot
try {
    $listener = Get-NetTCPConnection -LocalAddress $ListenAddress -LocalPort $ListenPort -State Listen -ErrorAction SilentlyContinue | Select-Object -First 1
    if (-not $listener) {
        Write-Host "No listener found on $ListenAddress`:$ListenPort. Nothing to stop."
        exit 0
    }

    $process = Get-ProcessInfo -ProcessId $listener.OwningProcess
    Write-Host "Found listener on $ListenAddress`:$ListenPort."
    Show-ProcessDetails $process

    if (-not (Test-IsConfirmedNuvioCmsProcess $process)) {
        Write-Warning "Listener is not confirmed as local NuvioCMS. Not stopping it."
        exit 1
    }

    $stopRootProcessId = Get-NuvioCmsStopRootProcessId -ProcessId $process.ProcessId
    if (-not $stopRootProcessId) {
        Write-Warning "Could not determine safe NuvioCMS process tree root. Not stopping anything."
        exit 1
    }

    $stopRoot = Get-ProcessInfo -ProcessId $stopRootProcessId
    Write-Host "Confirmed NuvioCMS stop root:"
    Show-ProcessDetails $stopRoot
    Stop-ProcessTree -ProcessId $stopRootProcessId

    Start-Sleep -Milliseconds 500
    $remaining = Get-NetTCPConnection -LocalAddress $ListenAddress -LocalPort $ListenPort -State Listen -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($remaining) {
        Write-Warning "Port $ListenAddress`:$ListenPort is still occupied by PID $($remaining.OwningProcess)."
        exit 1
    }

    Write-Host "Port $ListenAddress`:$ListenPort is free."
} finally {
    Pop-Location
}