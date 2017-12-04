package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	hclog "github.com/hashicorp/go-hclog"
	stan "github.com/nats-io/go-nats-streaming"
	"github.com/nicholasjackson/faas-nats/client"
	"github.com/nicholasjackson/faas-nats/config"
	"github.com/nicholasjackson/faas-nats/worker"
	yaml "gopkg.in/yaml.v2"
)

const appName = "faas_nats"

var configFile = flag.String("config", "", "configuration file continaing events to monitor")
var nc stan.Conn
var stats *statsd.Client
var logger hclog.Logger

func main() {
	fmt.Println("Starting OpenFaaS Queue (NATS.io)")

	flag.Parse()

	c := loadConfig()
	logger = setupLogging(c, appName)
	stats = setupStatsD(c.StatsD, appName)

	var err error
	nc, err = setupNats(c, appName)
	if err != nil {
		panic(err)
	}

	defer nc.Close()

	client := client.NewClient(
		c.Gateway,
		stats,
		logger.Named("gateway-client"),
	)

	worker := worker.NewNatsWorker(
		nc,
		client,
		stats,
		logger.Named("event-worker"),
	)
	worker.RegisterMessageListeners(c)

	http.DefaultServeMux.HandleFunc("/health", healthCheck)
	http.ListenAndServe(":9999", nil)
}

func loadConfig() config.Config {
	data, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatal("Config file does not exist:", err)
	}

	c := config.Config{}

	err = yaml.Unmarshal([]byte(data), &c)
	if err != nil {
		log.Fatal("Unable to read config", err)
	}

	fmt.Printf("Loaded config: %#s\n", c)

	return c
}

func setupLogging(c config.Config, appName string) hclog.Logger {
	logJSON := false
	if c.LogFormat == "json" {
		logJSON = true
	}

	appLogger := hclog.New(&hclog.LoggerOptions{
		Name:       appName,
		Level:      hclog.LevelFromString(c.LogLevel),
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

func setupNats(c config.Config, appName string) (stan.Conn, error) {
	clientID := fmt.Sprintf("%s-%d", appName, time.Now().UnixNano())
	nc, err := stan.Connect(c.NatsClusterID, clientID, stan.NatsURL(c.Nats))
	if err != nil {
		stats.Incr("connection.nats.failed", nil, 1)
		logger.Error("Unable to connect to nats server", "error", err)
	}

	stats.Incr("connection.nats.success", nil, 1)

	return nc, err
}

func healthCheck(rw http.ResponseWriter, r *http.Request) {
	if !nc.NatsConn().IsConnected() {
		stats.Incr("connection.nats.disconnected", nil, 1)

		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Sprint(rw, `{"nats": "not connected"}`)
	}
}
