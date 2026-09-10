# ZeroDrop

> **Passerelle d'Ingestion de Webhooks Haute Performance, Moteur de Rejeu & Tunneling Local.**  
> 100% Auto-Heberge (Self-Hosted), Binaire Unique Autonome en Go, SQLite (mode WAL), Hub WebSockets et Dashboard React 19 embarque.

[![Version](https://img.shields.io/badge/ZeroDrop-v0.1.0-blueviolet?style=for-the-badge)](https://github.com/nosleepman1/zerodrop)
[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=for-the-badge&logo=go)](https://golang.org)
[![Tests](https://img.shields.io/badge/Tests-100%25%20Passed-00C853?style=for-the-badge)](https://github.com/nosleepman1/zerodrop)
[![Self-Hosted CI/CD](https://img.shields.io/badge/CI%2FCD-Self--Hosted%20VM-FF6D00?style=for-the-badge)](https://github.com/nosleepman1/zerodrop)
[![License](https://img.shields.io/badge/License-MIT-black?style=for-the-badge)](LICENSE)

---

## Sommaire

- [1. Presentation & Philosophie](#1-presentation--philosophie)
- [2. Architecture Technique & Flux de Donnees](#2-architecture-technique--flux-de-donnees)
- [3. Demarrage Rapide](#3-demarrage-rapide)
- [4. Reference de l'API REST & Ingestion](#4-reference-de-lapi-rest--ingestion)
- [5. Moteur Cryptographique & Signatures HMAC](#5-moteur-cryptographique--signatures-hmac)
- [6. Reference du CLI ZeroDrop](#6-reference-du-cli-zerodrop)
- [7. Guide de Deploiement & Self-Hosting](#7-guide-de-deploiement--self-hosting)
- [8. CI/CD sur Self-Hosted Runner (VM Privee)](#8-cicd-sur-self-hosted-runner-vm-privee)
- [9. Cross-Compilation Multi-OS](#9-cross-compilation-multi-os)
- [10. Strategie de Tests & Assurance Qualite](#10-strategie-de-tests--assurance-qualite)

---

## 1. Presentation & Philosophie

**ZeroDrop** est une infrastructure moderne conçue pour resoudre l'enfer de l'integration, du debug et du monitoring de webhooks (Stripe, GitHub, Shopify, Slack, Clerk, Resend...) sans dependre de solutions SaaS tierces couteuses :

- **Zéro Perte de Donnees :** Reçoit et persiste chaque octet de requete avec acquittement instantane `202 Accepted` (< 2ms).
- **Zéro CGO / Binaire Autonome (< 12 Mo) :** Distribue le backend Go, le moteur SQLite WAL et le Dashboard Web React/Tailwind/Monaco dans un fichier executable unique.
- **Zéro Abonnement Cloud :** Deploiement en 1 clic sur n'importe quel VPS ou conteneur Docker personnel.
- **Relais Local Ultra-Rapide :** Tunneling WebSocket bidirectionnel (`zerodrop listen`) vers votre environnement de developpement local (`localhost:3000`).
- **Rejeu Interactif (Replay Studio) :** Rejouez ou mutez n'importe quel payload avec calcul automatique des delais et historique des reponses.

---

## 2. Architecture Technique & Flux de Donnees

```text
[ Emetteur Tiers ] (Stripe, GitHub, Shopify...)
        │
        ▼ POST /in/{endpoint_slug}
┌─────────────────────────────────────────────────────────────┐
│ 1. Passerelle d'Ingestion HTTP (Chi / net/http)             │
│    - Reponse immediate 202 Accepted (< 2ms)                 │
│    - Verification cryptographique HMAC (hmac.Equal)         │
│    - Capture fidele des Headers et du corps brut (Raw Body) │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. Persistance Relationnelle SQLite WAL (Pur Go)            │
│    - Mode WAL (Write-Ahead Logging) non-bloquant            │
│    - Indexation par slug, date et methode HTTP              │
└──────────────┬───────────────────────────────┬──────────────┘
               │                               │
               ▼ (Flux WebSockets)             ▼ (Tunnel WSS)
┌──────────────────────────────┐ ┌────────────────────────────┐
│ 3. Dashboard Web Embarque    │ │ 4. Agent CLI (zerodrop)    │
│  - Flux en direct temps reel │ │  - Relais vers localhost   │
│  - Editeur Monaco JSON       │ │  - Journalisation console  │
│  - Replay Studio & Diffs     │ └─────────────┬──────────────┘
│  - Export cURL & Types TS    │               │
└──────────────────────────────┘               ▼
                                   [ Application Locale Dev ]
                                      (localhost:3000/api)
```

---

## 3. Demarrage Rapide

### Option A : Execution du binaire pre-compile
```bash
# Lancement du serveur et du Dashboard Web sur http://localhost:8080
./bin/zerodrop.exe serve -port 8080
```

### Option B : Execution avec Go
```bash
go run . serve -port 8080
```

### Option C : Deploiement avec Docker Compose
```bash
docker compose up -d
```

---

## 4. Reference de l'API REST & Ingestion

### A. Ingestion Publique de Webhooks

| Methode | Route | Description | Reponse |
| :--- | :--- | :--- | :--- |
| `POST / PUT / PATCH` | `/in/{slug}` | Point d'entree d'ingestion | `202 Accepted` |

**Exemple de requete d'ingestion :**
```bash
curl -X POST "http://localhost:8080/in/default" \
  -H "Content-Type: application/json" \
  -d '{"event": "payment_intent.succeeded", "amount": 4200}'
```

**Reponse JSON retournee :**
```json
{
  "status": "accepted",
  "id": "req_9fcec545-7ac",
  "endpoint": "default",
  "signature_valid": null,
  "received_at": "2026-09-10T03:38:36Z"
}
```

---

### B. Gestion des Endpoints (`/api/endpoints`)

- `GET /api/endpoints` : Liste l'ensemble des endpoints configures.
- `POST /api/endpoints` : Cree un nouvel endpoint.
  ```json
  {
    "name": "Stripe Production",
    "slug": "stripe-prod",
    "provider": "stripe",
    "secret": "whsec_123456789",
    "forward_url": "http://localhost:3000/api/stripe"
  }
  ```
- `GET /api/endpoints/{id}` : Retourne les details d'un endpoint.
- `DELETE /api/endpoints/{id}` : Supprime un endpoint et ses requetes en cascade.

---

### C. Inspection & Rejeu des Requetes (`/api/requests`)

- `GET /api/requests?endpoint_id={id}&search={term}&limit=50&offset=0` : Liste les requetes avec pagination et recherche.
- `GET /api/requests/{id}` : Details complets avec corps brut non tronque.
- `POST /api/requests/{id}/replay` : Declenche un rejeu HTTP immediat.
  ```json
  {
    "target_url": "http://localhost:3000/api/webhook",
    "modified_body": "{\"event\": \"payment_intent.succeeded\", \"amount\": 9900}"
  }
  ```
- `GET /api/requests/{id}/replays` : Historique des rejeux et metriques de latence.
- `DELETE /api/requests` : Purge les requetes stockees.

---

## 5. Moteur Cryptographique & Signatures HMAC

ZeroDrop prend en charge la validation cryptographique stricte au format RFC 2104 avec comparaison en temps constant (`crypto/hmac.Equal`) :

| Fournisseur | Format d'en-tete | Algorithme & Structure |
| :--- | :--- | :--- |
| **Stripe** | `Stripe-Signature` | `t={timestamp},v1={HMAC-SHA256}` avec fenetre anti-rejeu (5 min) |
| **GitHub** | `X-Hub-Signature-256` | `sha256={HMAC-SHA256}` |
| **Shopify** | `X-Shopify-Hmac-Sha256` | `HMAC-SHA256` encode en Base64 standard |
| **Slack** | `X-Slack-Signature` | `v0={HMAC-SHA256}` sur `v0:{timestamp}:{body}` |
| **Generique** | `X-Signature`, `Signature` | `HMAC-SHA256` hexadecimal direct |

---

## 6. Reference du CLI ZeroDrop

Le binaire unifie `zerodrop` propose 4 commandes :

### 1. `zerodrop serve`
Demarre le serveur d'ingestion, le Hub WebSockets et sert le Dashboard Web embarque.
```bash
zerodrop serve -port 8080 -db /data/zerodrop.db
```

### 2. `zerodrop listen`
Demarre l'agent de tunneling local pour relayer les requetes du serveur cloud vers votre machine.
```bash
zerodrop listen --forward-to http://localhost:3000/api/webhook --server ws://webhooks.domaine.com
```

### 3. `zerodrop trigger`
Emet un webhook simulé avec calcul automatique de la signature cryptographique.
```bash
# Webhook Stripe
zerodrop trigger --provider stripe --event payment_intent.succeeded --secret whsec_test

# Webhook GitHub
zerodrop trigger --provider github --event pull_request.opened --secret mon_secret_github
```

### 4. `zerodrop version`
Affiche le numero de version et les informations de build.

---

## 7. Guide de Deploiement & Self-Hosting

### A. Deploiement Docker Compose (Production)
```yaml
# docker-compose.yml
services:
  zerodrop:
    build: .
    container_name: zerodrop
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
    environment:
      - TZ=UTC
```

### B. HTTPS Automatique avec Caddy (Recommande)
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

### C. Service Systemd (Linux VPS)
```ini
# /etc/systemd/system/zerodrop.service
[Unit]
Description=ZeroDrop Webhook Hub & Replay Engine
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/var/www/zerodrop
ExecStart=/usr/local/bin/zerodrop serve -port 8080 -db /var/www/zerodrop/zerodrop.db
Restart=always

[Install]
WantedBy=multi-user.target
```

---

## 8. CI/CD sur Self-Hosted Runner (VM Privee)

L'integralite de la chaîne d'integration continue et de publication de releases s'execute sur votre **propre machine virtuelle (Self-Hosted Runner)** sans consommer de quota payant GitHub Actions (`runs-on: self-hosted`).

### Installation du Runner sur votre VM Linux en 3 etapes :

1. **Executer le script d'installation automatique :**
   ```bash
   chmod +x deploy/setup-self-hosted-runner.sh
   ./deploy/setup-self-hosted-runner.sh
   ```

2. **Rattacher le runner a votre repository :**
   Rendez-vous dans les parametres de votre repository GitHub :  
   `Settings` -> `Actions` -> `Runners` -> `New self-hosted runner`  
   Recuperez le jeton de configuration, puis executez :
   ```bash
   cd ~/actions-runner
   ./config.sh --url https://github.com/nosleepman1/zerodrop --token <VOTRE_TOKEN>
   ```

3. **Demarrer le runner en service d'arriere-plan permanent :**
   ```bash
   sudo ./svc.sh install
   sudo ./svc.sh start
   ```

> Les workflows `.github/workflows/ci.yml` (tests automatiques sur push/PR) et `.github/workflows/release.yml` (publication de releases avec binaires sur les tags `v*`) s'executeront directement sur votre VM !

---

## 9. Cross-Compilation Multi-OS

Generez l'ensemble des binaires natifs en local sans consommer de minutes GitHub Actions :

```powershell
# Windows (PowerShell) :
.\scripts\build-all.ps1

# Linux / macOS (Bash) :
./scripts/build-all.sh
```

Binaires generes dans `./bin/` :
- `bin/zerodrop-windows-amd64.exe` (Windows 64-bit)
- `bin/zerodrop-linux-amd64` (Linux x86_64)
- `bin/zerodrop-linux-arm64` (Linux ARM64 / Raspberry Pi)
- `bin/zerodrop-darwin-arm64` (macOS Apple Silicon M1/M2/M3/M4)
- `bin/zerodrop-darwin-amd64` (macOS Intel)

---

## 10. Strategie de Tests & Assurance Qualite

L'ensemble des modules fait l'objet d'une couverture de tests automatisee :

```bash
# Execution de la suite complete
go test ./... -v
```

- `internal/api` : Tests d'integration HTTP (Ingestion, CRUD, Rejeu, Healthcheck).
- `internal/database` : Tests de transactions concurrentes, migrations et suppressions en cascade.
- `internal/hub` : Tests du modèle concurrent Goroutines/Channels et retransmissions WebSockets.
- `internal/replay` : Tests du moteur de rejeu avec serveurs HTTP fictifs et gestion des erreurs reseau.
- `internal/security` : Tests exhaustifs des validateurs HMAC et protections anti-rejeu.
- `internal/tunnel` : Tests de generation des charges utiles et calcul des signatures.

---

## Licence

Projet open-source distribue sous licence MIT. Developpe par [nosleepman1](https://github.com/nosleepman1).
