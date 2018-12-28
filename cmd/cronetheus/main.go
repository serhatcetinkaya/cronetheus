package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/serhatck/cronetheus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	configFile := flag.String("config", "config.yaml", "The Cronetheus config file")
	port := flag.String("port", ":9375", "Address on which to expose metrics and web interface.")

	// -alsologtostderr is true by default
	if alsoLogToStderr := flag.Lookup("alsologtostderr"); alsoLogToStderr != nil {
		alsoLogToStderr.DefValue = "true"
		alsoLogToStderr.Value.Set("true")
	}
	flag.Parse()

	c := &cronetheus.Config{}
	err := c.Init(*configFile)
	if err != nil {
		glog.Errorf("Fatal error: %q", err)
		os.Exit(2)
	}

	cron, _ := Schedule(c)
	cron.Start()

	go func() {
		sighup := make(chan os.Signal, 1)
		signal.Notify(sighup, syscall.SIGHUP)
		for {
			<-sighup
			glog.Infof("Reloading config")
			cron.Stop()
			c.Init(*configFile)
			cron, _ = Schedule(c)
			cron.Start()
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/config", ConfigHandlerFunc(c))
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { http.Error(w, "OK", http.StatusOK) })

	http.ListenAndServe(*port, nil)
}

// ConfigHandlerFunc is the HTTP handler for the `/config` page. It outputs the configuration marshaled in YAML format.
func ConfigHandlerFunc(config *cronetheus.Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s\n"+config.String(), "cron_config:")
	}
}
