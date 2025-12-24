package main

import (
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

func main() {
	var pormPort = flag.Int("porm.port", 9150, "port to expose prometheus metrics")
	flag.Parse()

	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector())

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	port := *pormPort
	log.Printf("Listening on port %d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		log.Fatalf("cannot start exporter: %s", err)
	}
}
