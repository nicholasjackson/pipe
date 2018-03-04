package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	stan "github.com/nats-io/go-nats-streaming"
)

var natsConnection = flag.String("nats", "nats://localhost:4222", "connection string for nats server")

func main() {
	flag.Parse()

	eventName := "example.echo"

	clientID := fmt.Sprintf("server-%d", time.Now().UnixNano())
	nc, err := stan.Connect("test-cluster", clientID, stan.NatsURL(*natsConnection))
	if err != nil {
		log.Fatal("Unable to connect to nats server: ", err)
	}

	nc.Subscribe("example.info.success", func(m *stan.Msg) {
		log.Println(string(m.Data))
	})

	log.Println("Sending event:", eventName)

	err = nc.Publish(eventName, []byte(`{"subject": "test"}`))
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(1 * time.Second)
}
