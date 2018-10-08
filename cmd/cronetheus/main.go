package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/serhatck/cronetheus"
	"net/http"
	"os"
)

func main() {
	configFile := flag.String("config", "config.yaml", "The Cronetheus config file")
	port := flag.String("port", ":9375", "Address on which to expose metrics and web interface. (defaults to \":9375\")")

	// -alsologtostderr is true by default
	if alsoLogToStderr := flag.Lookup("alsologtostderr"); alsoLogToStderr != nil {
		alsoLogToStderr.DefValue = "true"
		alsoLogToStderr.Value.Set("true")
	}
	flag.Parse()

	http.Handle("/metrics", promhttp.Handler())

	c, err := cronetheus.LoadConfigFile(*configFile)
	if err != nil {
		glog.Errorf("Fatal error: %q", err)
		os.Exit(2)
	}

	cron, _ := Schedule(c)
	cron.Start()
	http.ListenAndServe(*port, nil)
}
