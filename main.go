package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	stan "github.com/nats-io/go-nats-streaming"
	"github.com/nicholasjackson/faas-nats/client"
	"github.com/nicholasjackson/faas-nats/config"
	"github.com/nicholasjackson/faas-nats/worker"
	yaml "gopkg.in/yaml.v2"
)

var configFile = flag.String("config", "", "configuration file continaing events to monitor")
var nc stan.Conn

func main() {
	fmt.Println("Starting OpenFaaS Queue (NATS.io)")

	flag.Parse()

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

	clientID := fmt.Sprintf("server-%d", time.Now().UnixNano())
	nc, err = stan.Connect("faas-nats", clientID, stan.NatsURL(c.Nats))
	if err != nil {
		log.Fatal("Unable to connect to nats server")
	}
	defer nc.Close()

	client := client.NewClient(c.Gateway)
	worker := worker.NewNatsWorker(nc, client)
	worker.RegisterMessageListeners(c)

	http.DefaultServeMux.HandleFunc("/health", healthCheck)
	http.ListenAndServe(":9999", nil)
}

func healthCheck(rw http.ResponseWriter, r *http.Request) {
	if !nc.NatsConn().IsConnected() {
		fmt.Sprint(rw, `{"nats": "not connected"}`)
		rw.WriteHeader(http.StatusInternalServerError)
	}
}
