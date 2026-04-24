$ErrorActionPreference = "Stop"

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

Push-Location $PSScriptRoot
try {
    Import-EnvFile ".env"
    Import-EnvFile "ui/.env"

    Push-Location "ui"
    try {
        npm run build
        if ($LASTEXITCODE -ne 0) {
            exit $LASTEXITCODE
        }
    } finally {
        Pop-Location
    }

    go run ./examples/base serve
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
} finally {
    Pop-Location
}
