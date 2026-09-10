# ========================================================
# Étape 1 : Build de l'interface Web (React + Tailwind)
# ========================================================
FROM node:22-alpine AS ui-builder
WORKDIR /app/ui

# Installation des dépendances npm
COPY ui/package.json ui/package-lock.json ./
RUN npm ci

# Copie des sources du frontend et compilation
COPY ui/ ./
RUN npm run build

# ========================================================
# Étape 2 : Compilation du binaire Go statique autonome
# ========================================================
FROM golang:1.26-alpine AS go-builder
WORKDIR /app

# Installation de git et des certificats racine
RUN apk add --no-cache git ca-certificates

# Téléchargement des modules Go
COPY go.mod go.sum ./
RUN go mod download

# Copie de tout le code source et des fichiers compilés du frontend
COPY . .
COPY --from=ui-builder /app/ui/dist ./ui/dist

# Compilation statique ultra-optimisée (sans CGO)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/zerodrop .

# ========================================================
# Étape 3 : Image de Production Ultra-Légère (< 25 Mo)
# ========================================================
FROM alpine:3.21 AS runner

RUN apk add --no-cache ca-certificates tzdata

# Création d'un utilisateur non-root pour la sécurité
RUN addgroup -S zerodrop && adduser -S zerodrop -G zerodrop

WORKDIR /app

# Création du point de montage persistant pour la base SQLite
RUN mkdir -p /data && chown -R zerodrop:zerodrop /data /app

# Copie du binaire compilé
COPY --from=go-builder --chown=zerodrop:zerodrop /app/bin/zerodrop /app/zerodrop

USER zerodrop

# Exposition du port d'ingestion et du Dashboard
EXPOSE 8080

# Volume persistant pour la base SQLite
VOLUME ["/data"]

# Variables d'environnement par défaut
ENV PORT=8080
ENV DB_PATH=/data/zerodrop.db

# Healthcheck natif
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

ENTRYPOINT ["/app/zerodrop", "serve", "-port", "8080", "-db", "/data/zerodrop.db"]
