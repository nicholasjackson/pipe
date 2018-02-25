package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/DataDog/datadog-go/statsd"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/nicholasjackson/pipe/config"
)

const appName = "pipe"

var configFolder = flag.String("config", "", "directory containing configuration files")
var statsDAddress = flag.String("statsd", "localhost:8125", "statsD server")
var logFormat = flag.String("log_format", "text", "log format json | text")
var logLevel = flag.String("log_level", "INFO", "log level INFO | DEBUG | ERROR | TRACE")

var stats *statsd.Client
var logger hclog.Logger

var version = "notset"

func main() {
	fmt.Println("Starting Pipe Version:", version)

	flag.Parse()

	//c := loadConfig()
	logger = setupLogging(*logFormat, *logLevel, appName)
	stats = setupStatsD(*statsDAddress, appName)

	http.DefaultServeMux.HandleFunc("/health", healthCheck)
	http.ListenAndServe(":9999", nil)
}

func loadConfig() config.Config {
	c, err := config.ParseFolder(*configFolder)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Loaded config: %#v\n", c)

	return c
}

func setupLogging(logFormat, logLevel, appName string) hclog.Logger {
	logJSON := false
	if logFormat == "json" {
		logJSON = true
	}

	appLogger := hclog.New(&hclog.LoggerOptions{
		Name:       appName,
		Level:      hclog.LevelFromString(logLevel),
		JSONFormat: logJSON,
	})

	return appLogger
}

func setupStatsD(server, appName string) *statsd.Client {
	stats, err := statsd.New(server)
	if err != nil {
		logger.Warn("Unable to create StatsD connection")
	}
	stats.Namespace = appName + "."

	return stats
}

func healthCheck(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprint(rw, "Need to implement health checks")
}
