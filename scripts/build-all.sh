#!/usr/bin/env bash
# Script de cross-compilation locale multi-OS pour ZeroDrop
set -e

echo "[INFO] [1/2] Build du frontend React..."
cd ui && npm run build && cd ..

echo "[INFO] [2/2] Compilation des binaires Go multi-plateformes..."
mkdir -p bin

targets=(
    "windows/amd64/bin/zerodrop-windows-amd64.exe"
    "linux/amd64/bin/zerodrop-linux-amd64"
    "linux/arm64/bin/zerodrop-linux-arm64"
    "darwin/arm64/bin/zerodrop-darwin-arm64"
    "darwin/amd64/bin/zerodrop-darwin-amd64"
)

for target in "${targets[@]}"; do
    IFS="/" read -r os arch output <<< "$target"
    echo "  -> Compilation pour $os/$arch..."
    CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -ldflags="-s -w" -o "$output" .
done

echo "[OK] Tous les binaires ont ete generes dans le dossier ./bin/ !"
ls -lh bin/
