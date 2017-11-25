package main

import (
	"log"
	"time"

	"github.com/nats-io/nats"
)

func main() {
	eventName := "example.echo"

	nc, err := nats.Connect("nats://192.168.1.113:4222")
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
