# Script de cross-compilation locale multi-OS pour ZeroDrop
# Génère les binaires autonomes sans consommer de quota GitHub Actions.

Write-Host "⚡ [1/2] Build du frontend React..." -ForegroundColor Cyan
Push-Location ui
npm run build
Pop-Location

Write-Host "⚡ [2/2] Compilation des binaires Go multi-plateformes..." -ForegroundColor Cyan
New-Item -ItemType Directory -Force -Path bin | Out-Null

$targets = @(
    @{ OS = "windows"; Arch = "amd64"; Output = "bin/zerodrop-windows-amd64.exe" },
    @{ OS = "linux";   Arch = "amd64"; Output = "bin/zerodrop-linux-amd64" },
    @{ OS = "linux";   Arch = "arm64"; Output = "bin/zerodrop-linux-arm64" },
    @{ OS = "darwin";  Arch = "arm64"; Output = "bin/zerodrop-darwin-arm64" },
    @{ OS = "darwin";  Arch = "amd64"; Output = "bin/zerodrop-darwin-amd64" }
)

foreach ($t in $targets) {
    Write-Host "  -> Compilation pour $($t.OS)/$($t.Arch)..." -ForegroundColor Yellow
    $env:GOOS = $t.OS
    $env:GOARCH = $t.Arch
    $env:CGO_ENABLED = "0"
    go build -ldflags="-s -w" -o $t.Output .
}

# Restaure les variables d'environnement
Remove-Item Env:GOOS -ErrorAction SilentlyContinue
Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue

Write-Host "✅ Tous les binaires ont été générés dans le dossier ./bin/ !" -ForegroundColor Green
Get-ChildItem -Path bin | Select-Object Name, @{Name="Taille (Mo)"; Expression={[math]::Round($_.Length/1MB, 2)}}
