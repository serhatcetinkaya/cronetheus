package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	failedCronJobs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "failedCronJobs",
			Help: "Failed cron jobs, by cronId and user",
		},
		[]string{"cronid", "user"},
	)
)

func init() {
	prometheus.MustRegister(failedCronJobs)
}
