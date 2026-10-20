package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	handler "github.com/cyberbrown/magi/api"
	"github.com/cyberbrown/magi/internal/collector"
	"github.com/cyberbrown/magi/internal/cpu"
)

var version = "dev"

func main() {
	port := envOr("MAGI_PORT", "8780")
	token := os.Getenv("MAGI_TOKEN")
	if token == "" {
		log.Fatal("MAGI_TOKEN environment variable is required")
	}

	handler.Version = version

	// Take initial CPU baseline
	cpu.Init()

	// Start background collector
	col := collector.New()
	col.Start(2 * time.Second)

	h := handler.NewHandlers(col)
	auth := func(next http.HandlerFunc) http.HandlerFunc {
		return handler.AuthMiddleware(token, next)
	}

	mux := http.NewServeMux()

	// Public
	mux.HandleFunc("GET /health", h.Health)

	// Authenticated
	mux.HandleFunc("GET /gpu", auth(h.GPU))
	mux.HandleFunc("GET /vram", auth(h.Vram))
	mux.HandleFunc("GET /cpu", auth(h.CPU))
	mux.HandleFunc("GET /memory", auth(h.Memory))
	mux.HandleFunc("GET /disk", auth(h.Disk))
	mux.HandleFunc("GET /services", auth(h.Services))
	mux.HandleFunc("GET /system", auth(h.System))

	server := handler.CORSMiddleware(mux)

	log.Printf("MAGI v%s starting on :%s", version, port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), server); err != nil {
		log.Fatal(err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
