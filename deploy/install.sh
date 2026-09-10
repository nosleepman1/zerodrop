#!/usr/bin/env bash
# =========================================================================
# ZeroDrop - Script d'Installation Automatique 1-Clic pour VPS Linux
# =========================================================================
# Support : Ubuntu 20.04/22.04/24.04, Debian 11/12
# Usage : sudo ./deploy/install.sh [nom-de-domaine]
# =========================================================================
set -e

# Couleurs pour l'affichage console
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}================================================================${NC}"
echo -e "${BLUE}  ZeroDrop - Installation Automatique sur VPS Linux${NC}"
echo -e "${BLUE}================================================================${NC}"

# 1. Vérification des privilèges root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}[ERROR] Ce script doit etre execute avec les privileges root (sudo).${NC}"
    echo "Exemple : sudo ./deploy/install.sh"
    exit 1
fi

# 2. Récupération du nom de domaine
DOMAIN="$1"
if [ -z "$DOMAIN" ]; then
    echo ""
    echo -e "${YELLOW}Veuillez entrer votre nom de domaine ou sous-domaine (ex: webhooks.mon-domaine.com) :${NC}"
    read -p "Domaine : " DOMAIN
fi

if [ -z "$DOMAIN" ]; then
    echo -e "${RED}[ERROR] Nom de domaine obligatoire.${NC}"
    exit 1
fi

echo -e "${GREEN}[OK] Domaine configure : ${DOMAIN}${NC}"

# 3. Choix du mode de déploiement
echo ""
echo "Choisissez votre mode de deploiement :"
echo "  1) Docker Compose + Caddy (Recommande - Isole et cle en main)"
echo "  2) Binaire Natif Go + Systemd + Caddy (Ultra-leger - Consommation < 15 Mo RAM)"
read -p "Choix [1/2] (defaut: 1) : " DEPLOY_MODE
DEPLOY_MODE=${DEPLOY_MODE:-1}

# 4. Installation des prérequis système
echo ""
echo -e "${BLUE}[INFO] Mise a jour du systeme et installation des paquets de base...${NC}"
apt-get update -y
apt-get install -y curl git ufw debian-keyring debian-archive-keyring apt-transport-https

# 5. Installation et configuration de Caddy (HTTPS Automatique)
echo -e "${BLUE}[INFO] Installation de Caddy (Gestionnaire SSL automatique)...${NC}"
if ! command -v caddy &> /dev/null; then
    curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg --yes
    curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list
    apt-get update -y
    apt-get install -y caddy
fi

# Configuration du Caddyfile
echo -e "${BLUE}[INFO] Configuration du reverse-proxy Caddy pour ${DOMAIN}...${NC}"
cat <<EOF > /etc/caddy/Caddyfile
${DOMAIN} {
    reverse_proxy 127.0.0.1:8080 {
        header_up Host {host}
        header_up X-Real-IP {remote_host}
        header_up X-Forwarded-For {remote_host}
        header_up X-Forwarded-Proto {scheme}
    }
    encode gzip zstd
}
EOF

systemctl reload caddy || systemctl restart caddy

# 6. Déploiement selon le mode choisi
if [ "$DEPLOY_MODE" = "1" ]; then
    # Mode Docker Compose
    echo -e "${BLUE}[INFO] Verification et installation de Docker...${NC}"
    if ! command -v docker &> /dev/null; then
        curl -fsSL https://get.docker.com | sh
    fi

    echo -e "${BLUE}[INFO] Lancement de ZeroDrop via Docker Compose...${NC}"
    mkdir -p data
    docker compose up -d --build

else
    # Mode Binaire Natif Go + Systemd
    echo -e "${BLUE}[INFO] Verification et compilation du binaire natif...${NC}"
    if ! command -v go &> /dev/null; then
        apt-get install -y golang-go nodejs npm
        cd ui && npm ci && npm run build && cd ..
    fi

    mkdir -p /var/www/zerodrop
    go build -ldflags="-s -w" -o /usr/local/bin/zerodrop .
    chmod +x /usr/local/bin/zerodrop

    # Configuration Systemd
    cat <<EOF > /etc/systemd/system/zerodrop.service
[Unit]
Description=ZeroDrop Webhook Hub & Replay Engine
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/var/www/zerodrop
ExecStart=/usr/local/bin/zerodrop serve -port 8080 -db /var/www/zerodrop/zerodrop.db
Restart=always
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF

    chown -R www-data:www-data /var/www/zerodrop
    systemctl daemon-reload
    systemctl enable --now zerodrop
fi

# 7. Configuration du Pare-feu UFW
echo -e "${BLUE}[INFO] Configuration du pare-feu (Ports 80, 443 et 22 autorises)...${NC}"
ufw allow 22/tcp || true
ufw allow 80/tcp || true
ufw allow 443/tcp || true

# 8. Affichage du résumé
echo ""
echo -e "${GREEN}================================================================${NC}"
echo -e "${GREEN}  ZeroDrop a ete installe et deploye avec succes !${NC}"
echo -e "${GREEN}================================================================${NC}"
echo -e "  Dashboard Web & Hub : ${BLUE}https://${DOMAIN}${NC}"
echo -e "  Ingestion Webhooks  : ${BLUE}https://${DOMAIN}/in/{slug}${NC}"
echo -e "  Endpoint WebSocket  : ${BLUE}wss://${DOMAIN}/ws/tunnel${NC}"
echo ""
echo -e "${YELLOW}Pour relayer les webhooks vers votre machine locale (CLI) :${NC}"
echo -e "  zerodrop listen --server wss://${DOMAIN} --forward-to http://localhost:3000/api/webhook"
echo ""
echo -e "${BLUE}================================================================${NC}"
