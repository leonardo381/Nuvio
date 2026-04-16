$ErrorActionPreference = "Stop"

Push-Location $PSScriptRoot
try {
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
