// Package main fournit le point d'entrée CLI unifié pour l'exécutable ZeroDrop.
// Il orchestre les sous-commandes 'serve', 'listen', 'trigger' et 'version'.
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
	"github.com/nosleepman1/zerodrop/internal/tunnel"
)

// Version correspond au numéro de version actuel de ZeroDrop.
const Version = "0.1.0"

func printBanner() {
	fmt.Println(`
   _____              ____                   
  /____/__  _________ / __ \_________  ____  
     / / _ \/ ___/ __ \ / / / ___/ __ \/ __ \ 
    / /  __/ /  / /_/ // /_/ / /  / /_/ / /_/ /
   /_/\___/_/   \____/_____/_/    \____/ .___/ 
                                      /_/      
   ZeroDrop - High-Performance Webhook Replay & Tunnel Engine`)
	fmt.Printf("   Version : %s | Licence MIT - nosleepman1\n\n", Version)
}

func printUsage() {
	printBanner()
	fmt.Println("Usage:")
	fmt.Println("  zerodrop [commande] [options]")
	fmt.Println("")
	fmt.Println("Commandes disponibles :")
	fmt.Println("  serve      Demarre le serveur d'ingestion, le Hub WebSockets et le Dashboard (par defaut)")
	fmt.Println("  listen     Demarre l'agent de tunneling pour relayer les webhooks vers localhost")
	fmt.Println("  trigger    Emet un webhook de test simule (Stripe, GitHub, etc.) avec signature HMAC")
	fmt.Println("  version    Affiche la version actuelle de ZeroDrop")
	fmt.Println("")
	fmt.Println("Exemples d'utilisation :")
	fmt.Println("  zerodrop serve -port 8080 -db /data/zerodrop.db")
	fmt.Println("  zerodrop listen -forward-to http://localhost:3000/api/webhook")
	fmt.Println("  zerodrop trigger -provider stripe -event payment_intent.succeeded")
	fmt.Println("")
}

func main() {
	if len(os.Args) < 2 {
		runServe(os.Args[1:])
		return
	}

	cmd := os.Args[1]
	switch cmd {
	case "serve":
		runServe(os.Args[2:])
	case "listen":
		runListen(os.Args[2:])
	case "trigger":
		runTrigger(os.Args[2:])
	case "version", "-v", "--version":
		printBanner()
	case "help", "-h", "--help":
		printUsage()
	default:
		if cmd[0] == '-' {
			runServe(os.Args[1:])
		} else {
			fmt.Printf("Commande inconnue : '%s'\n\n", cmd)
			printUsage()
			os.Exit(1)
		}
	}
}

func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.String("port", "8080", "Port d'ecoute du serveur HTTP")
	dbPath := fs.String("db", "zerodrop.db", "Chemin d'acces vers la base de donnees SQLite")
	_ = fs.Parse(args)

	printBanner()
	log.Println("[INFO] Initialisation de ZeroDrop Core Server...")

	db, err := database.New(*dbPath)
	if err != nil {
		log.Fatalf("[FATAL] Erreur lors de l'initialisation de SQLite : %v", err)
	}
	defer db.Close()
	log.Printf("[INFO] Base de donnees SQLite prete (%s - Mode WAL)", *dbPath)

	eventHub := hub.NewHub()
	go eventHub.Run()
	log.Println("[INFO] Hub WebSockets temps reel actif")

	replayEngine := replay.NewEngine(db, eventHub)
	log.Println("[INFO] Moteur de Replay HTTP initialise")

	router := api.NewRouter(db, eventHub, replayEngine)

	server := &http.Server{
		Addr:         ":" + *port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		fmt.Printf("\n[INFO] Dashboard Web & Hub actifs sur : http://localhost:%s\n", *port)
		fmt.Printf("[INFO] URL d'ingestion des webhooks   : http://localhost:%s/in/{slug}\n", *port)
		fmt.Printf("[INFO] Endpoint WebSocket du Tunnel    : ws://localhost:%s/ws/tunnel\n\n", *port)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[FATAL] Erreur du serveur HTTP : %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[INFO] Arret gracieux du serveur ZeroDrop...")
}

func runListen(args []string) {
	fs := flag.NewFlagSet("listen", flag.ExitOnError)
	forwardTo := fs.String("forward-to", "http://localhost:3000/api/webhook", "URL locale vers laquelle relayer les requetes")
	serverURL := fs.String("server", "ws://localhost:8080", "URL du serveur ZeroDrop")
	endpoint := fs.String("endpoint", "*", "Slug de l'endpoint a ecouter (* pour tous)")
	_ = fs.Parse(args)

	agent := tunnel.NewAgent(tunnel.Config{
		ServerURL:    *serverURL,
		EndpointSlug: *endpoint,
		ForwardTo:    *forwardTo,
	})

	if err := agent.Start(); err != nil {
		log.Fatalf("[FATAL] Erreur de l'agent de tunneling : %v", err)
	}
}

func runTrigger(args []string) {
	fs := flag.NewFlagSet("trigger", flag.ExitOnError)
	targetURL := fs.String("url", "http://localhost:8080/in/default", "URL d'ingestion ZeroDrop")
	provider := fs.String("provider", "stripe", "Fournisseur de webhook (stripe, github, custom)")
	event := fs.String("event", "payment_intent.succeeded", "Nom de l'evenement a simuler")
	secret := fs.String("secret", "", "Secret HMAC optionnel pour signer le payload")
	_ = fs.Parse(args)

	err := tunnel.TriggerSampleWebhook(tunnel.TriggerConfig{
		TargetURL: *targetURL,
		Provider:  *provider,
		EventName: *event,
		Secret:    *secret,
	})
	if err != nil {
		log.Fatalf("[FATAL] Erreur lors de l'emission du webhook : %v", err)
	}
}
