package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Load configuration
	config := LoadConfig()

	// Create and register collector
	collector := NewNextcloudCollector(config)
	prometheus.MustRegister(collector)

	// Setup HTTP server
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
<head><title>Nextcloud Exporter</title></head>
<body>
<h1>Nextcloud Exporter</h1>
<p><a href="/metrics">Metrics</a></p>
</body>
</html>`))
	})

	log.Printf("Starting Nextcloud exporter on %s", config.ListenAddr)
	log.Printf("Fetching metrics from: %s", config.BaseURL)
	log.Printf("Fetch interval: %s (to avoid rate limiting)", config.FetchInterval)
	if err := http.ListenAndServe(config.ListenAddr, nil); err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}
