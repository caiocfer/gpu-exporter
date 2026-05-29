package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/caiocfer/gpu-exporter/internal/collector"
	"github.com/caiocfer/gpu-exporter/internal/nvml"
	"github.com/caiocfer/gpu-exporter/internal/proc"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	addr := flag.String("addr", ":9835", "listen address")
	path := flag.String("path", "/metrics", "metrics path")
	cacheTTL := flag.Duration("cache-ttl", 60*time.Second, "process metadata cache TTL")
	procRoot := flag.String("proc", "/proc", "path to proc filesystem")
	flag.Parse()

	proc.Root = *procRoot

	if err := nvml.Init(); err != nil {
		log.Fatalf("NVML init: %v", err)
	}
	defer nvml.Shutdown()

	reg, err := collector.NewRegistry(*cacheTTL)
	if err != nil {
		log.Fatalf("registry: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle(*path, promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	log.Printf("listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}
