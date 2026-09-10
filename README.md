# ? ZeroDrop

> **High-Performance Webhook Ingestion, Real-Time Replay Engine & Local Tunnel Gateway.**  
> Built with Go, SQLite (WAL mode), WebSockets, and a modern embedded React Dashboard.

![ZeroDrop Banner](https://img.shields.io/badge/ZeroDrop-v0.1.0-blueviolet?style=for-the-badge)
![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=for-the-badge&logo=go)
![Architecture](https://img.shields.io/badge/Zero--Drop-100%25%20Guaranteed-00C853?style=for-the-badge)
![License](https://img.shields.io/badge/License-MIT-black?style=for-the-badge)

---

## ?? Key Features

- **? Zero Data Loss Ingestion:** Ultra-fast HTTP ingestion gateway capable of handling bursts with instant 202 responses.
- **?? Interactive Replay Studio:** Replay any captured webhook with a single click, modify payloads/headers, and compare diffs.
- **?? Real-Time Local Tunneling:** Forward cloud webhooks directly to your \localhost\ services via secure WebSockets.
- **??? Cryptographic Signature Verification:** Built-in verification for Stripe (HMAC-SHA256), GitHub (HMAC-SHA256), Shopify, and custom secrets.
- **?? Single Binary Distribution:** Embedded React/Tailwind/Monaco web dashboard inside the compiled Go binary. Zero external runtime dependencies.
- **?? Embedded Storage:** Powered by high-speed SQLite with Write-Ahead Logging (WAL) for instant search and persistence.

---

## ??? Architecture Overview

\\\
[ Webhook Provider ] (Stripe, GitHub, Shopify, etc.)
        ¦
        ? POST /in/{endpoint_slug}
+-------------------------------------------------------------+
¦ ZeroDrop Ingestion Gateway (Go Engine)                      ¦
¦  - Instant 202 Accepted                                     ¦
¦  - Raw Body & Header Preservation                           ¦
¦  - Cryptographic Signature Verification                     ¦
+-------------------------------------------------------------+
                       ¦
                       ?
+-------------------------------------------------------------+
¦ Zero-Loss Storage (SQLite WAL Mode)                         ¦
+-------------------------------------------------------------+
               ¦                               ¦
               ? (Live WebSockets / SSE)       ? (WSS Tunnel)
+------------------------------+ +----------------------------+
¦ Embedded Web Dashboard (UI)  ¦ ¦ ZeroDrop CLI Agent         ¦
¦  - Real-Time Event Stream    ¦ ¦  - Forward to localhost    ¦
¦  - Monaco JSON Editor        ¦ ¦  - Stream response logs    ¦
¦  - Payload Mutation & Replay ¦ +----------------------------+
¦  - Diff & Latency Inspection ¦               ¦
+------------------------------+               ?
                                   [ Local Dev Server ]
                                      (localhost:3000)
\\\

---

## ??? Tech Stack

- **Core & Backend:** Go (Golang), \gorilla/websocket\ / \
et/http\, \modernc.org/sqlite\
- **Frontend / Dashboard:** React 19, Vite, Tailwind CSS, Lucide Icons, Monaco Editor (embedded via \//go:embed\)
- **CLI / Tunneling Agent:** Go CLI with subcommands (\zerodrop serve\, \zerodrop listen\)

---

## ?? License

MIT © [nosleepman1](https://github.com/nosleepman1)
