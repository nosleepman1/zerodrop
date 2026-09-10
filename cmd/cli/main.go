package main

import (
	"flag"
	"log"

	"github.com/nosleepman1/zerodrop/internal/tunnel"
)

func main() {
	forwardTo := flag.String("forward-to", "http://localhost:3000/api/webhook", "URL locale vers laquelle relayer les requetes")
	serverURL := flag.String("server", "ws://localhost:8080", "URL du serveur ZeroDrop")
	endpoint := flag.String("endpoint", "*", "Slug de l'endpoint a ecouter")
	flag.Parse()

	agent := tunnel.NewAgent(tunnel.Config{
		ServerURL:    *serverURL,
		EndpointSlug: *endpoint,
		ForwardTo:    *forwardTo,
	})

	if err := agent.Start(); err != nil {
		log.Fatalf("[FATAL] Erreur agent de tunneling : %v", err)
	}
}
