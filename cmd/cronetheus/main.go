package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/serhatck/cronetheus"
	"os"
)

func main() {
	configFile := flag.String("config", "config.yaml", "The Cronetheus config file")

	// -alsologtostderr is true by default
	if alsoLogToStderr := flag.Lookup("alsologtostderr"); alsoLogToStderr != nil {
		alsoLogToStderr.DefValue = "true"
		alsoLogToStderr.Value.Set("true")
	}
	flag.Parse()

	c, err := cronetheus.LoadConfigFile(*configFile)
	if err != nil {
		glog.Errorf("Fatal error: %q", err)
		os.Exit(2)
	}

	cron, _ := Schedule(c)
	cron.Start()
	select {}
}
