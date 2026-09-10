#!/usr/bin/env bash
# =========================================================================
# Script d'installation et de configuration automatique du Self-Hosted Runner
# GitHub Actions pour ZeroDrop sur votre Machine Virtuelle (Linux)
# =========================================================================
set -e

echo "================================================================"
echo "[INFO] Installation du Runner GitHub Actions Self-Hosted"
echo "================================================================"

# 1. Mise à jour et prérequis
echo "[INFO] [1/4] Verification et installation des outils requis..."
sudo apt-get update -y || sudo yum update -y || true
sudo apt-get install -y curl tar git jq libicu-dev || sudo yum install -y curl tar git jq libicu || true

# 2. Création du répertoire d'accueil
RUNNER_DIR="$HOME/actions-runner"
mkdir -p "$RUNNER_DIR" && cd "$RUNNER_DIR"

# 3. Récupération de la dernière version du runner GitHub
echo "[INFO] [2/4] Telechargement du package GitHub Actions Runner..."
RUNNER_VERSION=$(curl -s https://api.github.com/repos/actions/runner/releases/latest | jq -r '.tag_name' | sed 's/v//')
if [ -z "$RUNNER_VERSION" ] || [ "$RUNNER_VERSION" = "null" ]; then
    RUNNER_VERSION="2.321.0"
fi

RUNNER_ARCH="x64"
if [ "$(uname -m)" = "aarch64" ]; then
    RUNNER_ARCH="arm64"
fi

RUNNER_FILE="actions-runner-linux-${RUNNER_ARCH}-${RUNNER_VERSION}.tar.gz"
RUNNER_URL="https://github.com/actions/runner/releases/download/v${RUNNER_VERSION}/${RUNNER_FILE}"

if [ ! -f "$RUNNER_FILE" ]; then
    curl -o "$RUNNER_FILE" -L "$RUNNER_URL"
    tar xzf "$RUNNER_FILE"
fi

# 4. Instructions de configuration
echo ""
echo "================================================================"
echo "[INFO] [3/4] Enregistrement du Runner aupres de votre repo"
echo "================================================================"
echo "Rendez-vous sur : https://github.com/nosleepman1/zerodrop/settings/actions/runners/new"
echo "Recuperez votre jeton de configuration (TOKEN), puis executez :"
echo ""
echo "  cd ~/actions-runner"
echo "  ./config.sh --url https://github.com/nosleepman1/zerodrop --token <VOTRE_TOKEN>"
echo ""
echo "================================================================"
echo "[INFO] [4/4] Installation en service Systemd (24h/24 en arriere-plan)"
echo "================================================================"
echo "Une fois ./config.sh execute, activez le demarrage automatique :"
echo ""
echo "  sudo ./svc.sh install"
echo "  sudo ./svc.sh start"
echo "  sudo ./svc.sh status"
echo ""
echo "[OK] Le runner tournera en continu sur votre VM avec un cout GitHub de $0 !"
