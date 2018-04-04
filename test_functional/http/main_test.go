package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/godog"
	stan "github.com/nats-io/go-nats-streaming"
	"github.com/nicholasjackson/pipe/providers"
	"github.com/nicholasjackson/pipe/server"
	"github.com/nicholasjackson/pipe/test_functional/helpers"
)

var natsClient stan.Conn
var myMessageChannel chan *providers.Message
var pipeServer *server.PipeServer
var subs stan.Subscription
var log *bytes.Buffer
var httpServer *http.Server

func TestMain(m *testing.M) {
	myMessageChannel = make(chan *providers.Message, 1)
	helpers.MainTest(m, FeatureContext)
}

func pipeIsRunningAndConfigured() error {
	return nil
}

func iCallAnHttpEndpoint() error {
	resp, err := http.Post("http://localhost:18091/", "application/json", nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	return nil
}

func iExpectPipeToMakeAnOutboundHttpCall() error {
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

		httpServer = &http.Server{
			Addr: "localhost:18092",
		}
		http.HandleFunc("/", serveHTTP)

		go func() {
			err := httpServer.ListenAndServe()

			if err != http.ErrServerClosed {
				log.WriteString("Error starting server:" + err.Error())
			}
		}()

		// wait for server to start
		time.Sleep(5 * time.Second)
	})

	s.AfterScenario(func(interface{}, error) {
		ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
		httpServer.Shutdown(ctx)

		fmt.Println("\nLog output:")
		l, _ := ioutil.ReadAll(log)
		fmt.Println(string(l))
	})

	s.Step(`^Pipe is running and configured$`, pipeIsRunningAndConfigured)
	s.Step(`^I call an http endpoint$`, iCallAnHttpEndpoint)
	s.Step(`^I expect pipe to make an outbound http call$`, iExpectPipeToMakeAnOutboundHttpCall)
}

func serveHTTP(rw http.ResponseWriter, r *http.Request) {
	msg := providers.Message{}

	myMessageChannel <- &msg
}
