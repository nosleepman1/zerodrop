package main

import (
	"flag"
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
	port := flag.String("port", "8080", "Port d'ecoute du serveur ZeroDrop")
	dbPath := flag.String("db", "zerodrop.db", "Chemin d'acces vers SQLite")
	flag.Parse()

	log.Println("[INFO] Initialisation de ZeroDrop Core Engine...")

	db, err := database.New(*dbPath)
	if err != nil {
		log.Fatalf("[FATAL] Erreur SQLite : %v", err)
	}
	defer db.Close()

	eventHub := hub.NewHub()
	go eventHub.Run()

	replayEngine := replay.NewEngine(db, eventHub)
	router := api.NewRouter(db, eventHub, replayEngine)

	server := &http.Server{
		Addr:         ":" + *port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("[INFO] Serveur actif sur http://localhost:%s", *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[FATAL] Erreur serveur : %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[INFO] Arret du serveur ZeroDrop.")
}
