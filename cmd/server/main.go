package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nosleepman1/zerodrop/internal/api"
	"github.com/nosleepman1/zerodrop/internal/database"
	"github.com/nosleepman1/zerodrop/internal/hub"
	"github.com/nosleepman1/zerodrop/internal/replay"
)

func main() {
	port := flag.String("port", "8080", "Port d'écoute du serveur ZeroDrop")
	dbPath := flag.String("db", "zerodrop.db", "Chemin vers le fichier de base de données SQLite")
	flag.Parse()

	log.Println("⚡ Initialisation de ZeroDrop Core Engine...")

	// 1. Initialisation de la base SQLite (WAL mode)
	db, err := database.New(*dbPath)
	if err != nil {
		log.Fatalf("❌ Erreur fatale SQLite : %v", err)
	}
	defer db.Close()
	log.Printf("💾 Base de données SQLite connectée avec succès (%s)", *dbPath)

	// 2. Démarrage du Hub WebSocket
	eventHub := hub.NewHub()
	go eventHub.Run()
	log.Println("🔄 Hub WebSockets temps réel démarré")

	// 3. Initialisation du Moteur de Replay
	replayEngine := replay.NewEngine(db, eventHub)
	log.Println("🎯 Moteur de Replay HTTP initialisé")

	// 4. Configuration du Routeur HTTP
	router := api.NewRouter(db, eventHub, replayEngine)

	server := &http.Server{
		Addr:         ":" + *port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 5. Gestion de l'arrêt gracieux (Graceful Shutdown)
	go func() {
		log.Printf("🚀 Serveur ZeroDrop actif sur http://localhost:%s", *port)
		log.Printf("📥 Ingestion webhook prête sur : http://localhost:%s/in/{slug}", *port)
		log.Printf("🌐 API REST active sur : http://localhost:%s/api", *port)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Erreur du serveur HTTP : %v", err)
		}
	}()

	// Attente du signal de terminaison (Ctrl+C, SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Arrêt gracieux de ZeroDrop...")
	fmt.Println("Au revoir !")
}
