package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var pingCounter = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "ping_request_counter",
	Help: "The total number of ping requests",
})

func ping(w http.ResponseWriter, req *http.Request) {
	pingCounter.Inc()
	_, err := fmt.Fprintf(w, "pong")
	if err != nil {
		return
	}
}

func main() {
	// prometheus.MustRegister(pingCounter)

	registry := prometheus.NewRegistry()
	registry.MustRegister(pingCounter)

	http.HandleFunc("/ping", ping)
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	err := http.ListenAndServe(":9095", nil)
	if err != nil {
		return
	}
}

func init() {
	prometheus.Unregister(prometheus.NewGoCollector())
	prometheus.Unregister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
}
