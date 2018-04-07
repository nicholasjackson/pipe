package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/godog"
	stan "github.com/nats-io/go-nats-streaming"
	"github.com/nicholasjackson/pipe/server"
	"github.com/nicholasjackson/pipe/test_functional/helpers"
)

var natsClient stan.Conn
var myMessageChannel chan *stan.Msg
var pipeServer *server.PipeServer
var subs stan.Subscription
var log *bytes.Buffer

func TestMain(m *testing.M) {
	myMessageChannel = make(chan *stan.Msg, 1)
	helpers.MainTest(m, FeatureContext)
}

func natsIsRunning() error {
	return nil
}

func iReceiveAMessage() error {
	err := natsClient.Publish("messagein", []byte("testdata"))
	return err
}

func iExpectAActionMessageToBePublished() error {
	select {
	case <-myMessageChannel:
		return nil
	case <-time.After(10 * time.Second):
		return fmt.Errorf("Timeout waiting for message")
	}

	return nil
}

func FeatureContext(s *godog.Suite) {
	s.BeforeScenario(func(interface{}) {
		var err error
		pipeServer, log, err = helpers.StartServer(".")
		if err != nil {
			panic(err)
		}

		natsurl := stan.NatsURL("nats://" + os.Getenv("nats_server") + ":4222")
		clientid := "abc123"

		natsClient, err = stan.Connect(os.Getenv("nats_cluster_id"), clientid, natsurl)
		if err != nil {
			panic(err)
		}

		subs, err = natsClient.Subscribe("messageout", func(msg *stan.Msg) {
			myMessageChannel <- msg
		})
		if err != nil {
			panic(err)
		}
	})

	s.AfterScenario(func(interface{}, error) {
		natsClient.Close()

		fmt.Println("\nLog output:")
		l, _ := ioutil.ReadAll(log)
		fmt.Println(string(l))
	})

	s.Step(`^Nats is running$`, natsIsRunning)
	s.Step(`^I receive a message$`, iReceiveAMessage)
	s.Step(`^I expect a action message to be published$`, iExpectAActionMessageToBePublished)
}
