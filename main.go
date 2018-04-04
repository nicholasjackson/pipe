package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/DataDog/datadog-go/statsd"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/nicholasjackson/pipe/config"
	"github.com/nicholasjackson/pipe/logger"
	"github.com/nicholasjackson/pipe/server"
)

const appName = "pipe"

var configFolder = flag.String("config", "", "directory containing configuration files")
var statsDAddress = flag.String("statsd", "localhost:8125", "statsD server")
var logFormat = flag.String("log_format", "text", "log format json | text")
var logLevel = flag.String("log_level", "INFO", "log level INFO | DEBUG | ERROR | TRACE")

var version = "notset"

func main() {
	fmt.Println("Starting Pipe Version:", version)

	flag.Parse()

	l := createLogger(*logFormat, *logLevel, "pipe", *statsDAddress)

	c := loadConfig(l)
	s := server.New(c, l)

	s.Listen()

	http.DefaultServeMux.HandleFunc("/health", healthCheck)
	http.ListenAndServe(":9999", nil)
}

func loadConfig(l logger.Logger) *config.Config {
	c, err := config.ParseFolder(*configFolder)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Loaded config: %#v\n", c)

	c.Pipes, err = config.SetupPipes(c, l)
	if err != nil {
		log.Fatal(err)
	}

	return c
}

func createLogger(logFormat, logLevel, appName, statsDAddress string) logger.Logger {
	l := setupLogging(logFormat, logLevel, appName)
	s := setupStatsD(statsDAddress, appName)
	return logger.New(l, s)
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
		log.Println("Unable to create StatsD connection")
	}
	stats.Namespace = appName + "."

	return stats
}

func healthCheck(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprint(rw, "Need to implement health checks")
}
