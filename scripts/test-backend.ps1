Param()

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$toolchainDir = Join-Path $repoRoot ".toolchains\go"
$toolchainZip = Join-Path $repoRoot ".toolchains\go1.26.2.windows-amd64.zip"
$goExe = Join-Path $toolchainDir "bin\go.exe"

if (!(Test-Path $goExe)) {
  if (!(Test-Path $toolchainZip)) {
    Write-Host "Downloading Go 1.26.2 toolchain zip..."
    curl.exe -L "https://go.dev/dl/go1.26.2.windows-amd64.zip" -o $toolchainZip
  }
  if (Test-Path $toolchainDir) {
    Remove-Item -Recurse -Force $toolchainDir
  }
  Write-Host "Extracting Go toolchain..."
  tar -xf $toolchainZip -C (Join-Path $repoRoot ".toolchains")
}

$env:GOROOT = $toolchainDir
$env:GOPATH = "C:\Users\adiya\go"
$env:GOPROXY = "https://proxy.golang.org,direct"
$env:GOSUMDB = "sum.golang.org"

Write-Host "Using: $(& $goExe version)"
Push-Location (Join-Path $repoRoot "backend")
try {
  & $goExe mod tidy
  & $goExe test ./...
} finally {
  Pop-Location
}
