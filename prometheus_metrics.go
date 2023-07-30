package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	sentRecordsTotalCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "nflow_generator_sent_records_total",
		Help: "The total number of sent netflow records",
	})
	sentRecordsTotalBytesCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "nflow_generator_sent_records_total_bytes",
		Help: "The total number of sent netflow record bytes",
	})
	sentNetflowTotalCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "nflow_generator_sent_netflow_total",
		Help: "The total number of sent netflow packets",
	})
)

const PROM_PORT = "2112"

func HandleMetricsServer() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":"+PROM_PORT, nil)
}
