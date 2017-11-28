package main

import (
	"flag"
	"log"
	"time"

	"github.com/nats-io/nats"
)

var natsConnection = flag.String("nats", "nats://localhost:4222", "connection string for nats server")

func main() {
	flag.Parse()

	eventName := "example.echo"

	nc, err := nats.Connect(*natsConnection)
	if err != nil {
		log.Fatal("Unable to connect to nats server")
	}

	nc.Subscribe("example.info.success", func(m *nats.Msg) {
		log.Println(string(m.Data))
	})

	log.Println("Sending event:", eventName)

	err = nc.Publish(eventName, []byte(`{"subject": "test"}`))
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(1 * time.Second)
}
