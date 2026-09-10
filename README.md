# ⚡ ZeroDrop

> **Passerelle d'Ingestion Webhook Haute Performance, Moteur de Rejeu & Tunneling Local.**  
> 100% Self-Hosted, Binaire Unique Autonome en Go, SQLite (mode WAL), Hub WebSockets et Dashboard React/Tailwind/Monaco embarqué.

![ZeroDrop Banner](https://img.shields.io/badge/ZeroDrop-v0.1.0-blueviolet?style=for-the-badge)
![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=for-the-badge&logo=go)
![Architecture](https://img.shields.io/badge/Zero--Drop-100%25%20Guaranteed-00C853?style=for-the-badge)
![License](https://img.shields.io/badge/License-MIT-black?style=for-the-badge)

---

## 🚀 Pourquoi ZeroDrop ?

1. **🔒 100% Auto-Hébergé (Self-Hosted & Zéro Coût Cloud) :**  
   Déployez votre propre instance sur n'importe quel VPS (OVH, Hetzner, DigitalOcean à 3€/mois) ou serveur local. Zéro abonnement SaaS tiers, maîtrise totale de vos données.
2. **⚡ Performance Brute & Haute Disponibilité :**  
   Réponse immédiate `202 Accepted` en < 2ms pour encaisser les pics de charge sans jamais bloquer l'émetteur (Stripe, GitHub, Shopify...).
3. **🛡️ Vérification Cryptographique HMAC :**  
   Validation automatique intégrée des signatures Stripe (`Stripe-Signature`), GitHub (`X-Hub-Signature-256`), Shopify, Slack et secrets personnalisés.
4. **📦 Binaire Unique Autonome (Single Binary < 12 Mo) :**  
   L'interface web React 19 / Monaco Editor est **embarquée directement dans le binaire Go**. Zéro dépendance Node.js sur votre serveur de production.
5. **🔄 Replay Studio Interactif :**  
   Rejouez n'importe quelle requête en 1 clic vers votre serveur local ou une URL distante, avec ou sans mutation de payload.
6. **🔌 Agent de Tunneling Local (`zerodrop listen`) :**  
   Relayez instantanément vos webhooks distants vers votre `localhost:3000` via une connexion WebSocket sécurisée et résiliente.

---

## 🏗️ Architecture Globale

```
[ Émetteur Externe ] (Stripe, GitHub, Shopify...)
        │
        ▼ POST /in/{endpoint_slug}
┌─────────────────────────────────────────────────────────────┐
│ 1. ZeroDrop Ingestion Gateway (Go Engine)                   │
│    - Réponse immédiate 202 Accepted (< 2ms)                 │
│    - Vérification cryptographique de signature HMAC         │
│    - Préservation intégrale des Headers & Body brut         │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. Base de Données SQLite Ultra-Rapide (Mode WAL Pur Go)    │
│    - Transactions concurrentes non-bloquantes               │
│    - Indexation et persistance zero-loss                    │
└──────────────┬───────────────────────────────┬──────────────┘
               │                               │
               ▼ (WebSockets / SSE)            ▼ (WSS Tunnel)
┌──────────────────────────────┐ ┌────────────────────────────┐
│ 3. Dashboard Web Embarqué    │ │ 4. ZeroDrop CLI Agent      │
│  - Flux en direct temps réel │ │  - Forward vers localhost  │
│  - Éditeur Monaco JSON       │ │  - Logs colorés en console │
│  - Replay & Diff Studio      │ └─────────────┬──────────────┘
│  - Export cURL & Types TS    │               │
└──────────────────────────────┘               ▼
                                   [ Application Locale Dev ]
                                      (localhost:3000/api)
```

---

## 🛠️ Guide de Déploiement Self-Hosted

### Option 1 : Déploiement 1-Clic avec Docker Compose (Recommandé)

Créez votre fichier `docker-compose.yml` :

```yaml
services:
  zerodrop:
    image: ghcr.io/nosleepman1/zerodrop:latest
    # ou build local :
    # build: .
    container_name: zerodrop
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
    environment:
      - TZ=UTC
```

Lancez le conteneur en arrière-plan :
```bash
docker compose up -d
```
> Vos données SQLite sont automatiquement conservées dans le dossier `./data/`.

---

### Option 2 : Binaire Natif & Service Systemd (Linux VPS)

1. Téléchargez ou compilez le binaire Linux :
   ```bash
   sudo cp bin/zerodrop-linux-amd64 /usr/local/bin/zerodrop
   sudo chmod +x /usr/local/bin/zerodrop
   ```

2. Installez le service `systemd` (inclus dans `deploy/zerodrop.service`) :
   ```bash
   sudo mkdir -p /var/www/zerodrop
   sudo cp deploy/zerodrop.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable --now zerodrop
   ```

---

### Option 3 : HTTPS Automatique avec Caddy (Recommandé en Production)

Pour recevoir de vrais webhooks depuis Stripe/GitHub, votre serveur doit être accessible en HTTPS.  
Avec **Caddy**, la génération et le renouvellement des certificats SSL Let's Encrypt sont 100% automatiques :

```caddy
# /etc/caddy/Caddyfile
webhooks.votre-domaine.com {
    reverse_proxy 127.0.0.1:8080 {
        header_up Host {host}
        header_up X-Real-IP {remote_host}
        header_up X-Forwarded-For {remote_host}
        header_up X-Forwarded-Proto {scheme}
    }
}
```

---

## 💻 Utilisation du CLI Développeur

### 1. Démarrer le serveur et le dashboard web
```bash
zerodrop serve -port 8080
```
Ouvrez ensuite votre navigateur sur **http://localhost:8080**.

### 2. Démarrer le tunnel local (Relais vers votre machine)
```bash
zerodrop listen --forward-to http://localhost:3000/api/webhook --server ws://webhooks.votre-domaine.com
```

### 3. Simuler un webhook de test en ligne de commande
```bash
zerodrop trigger --provider stripe --event payment_intent.succeeded
zerodrop trigger --provider github --event pull_request.opened
```

---

## 🔨 Cross-Compilation Locale Multi-OS (Zéro Quota CI)

Vous pouvez générer instantanément tous les binaires exécutables pour Windows, Linux et macOS en local sans consommer de minutes GitHub Actions :

```powershell
# Sous Windows (PowerShell) :
.\scripts\build-all.ps1

# Sous Linux / macOS (Bash) :
./scripts/build-all.sh
```

---

## 📄 Licence

Projet open-source distribué sous licence MIT © [nosleepman1](https://github.com/nosleepman1).
