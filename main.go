package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/webmakersteve/elastic_prom_recorder/config"
	"github.com/webmakersteve/elastic_prom_recorder/recorder"
	"net/http"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})

	http.Handle("/metrics", promhttp.Handler())

	serviceConfig, err := config.Load("/Users/stephen/go/src/github.com/webmakersteve/elastic_prom_recorder/example.yaml")

	if err != nil {
		log.Fatalf("Failed to read configuration: %s", err)
	}

	// Really want to make a list of the group runners
	
	for _, group := range serviceConfig.Groups {
		g, err := recorder.NewGroup(&group)

		if err != nil {
			log.Fatalf("Failed to initialize group: %s", err)
		}

		// Run this goroutine
		go recorder.Execute(&g, context.Background())
	}

	http.ListenAndServe(":8080", nil)
}
