# Nats.io message listener for [OpenFaaS](https://github.com/openfaas/faas)
[![Docker Repository on Quay](https://quay.io/repository/nicholasjackson/faas-nats/status "Docker Repository on Quay")](https://quay.io/repository/nicholasjackson/faas-nats)
[![CircleCI](https://circleci.com/gh/nicholasjackson/faas-nats.svg?style=svg)](https://circleci.com/gh/nicholasjackson/faas-nats)

This project allows you to listen to Nats.io messages and call OpenFaas functions.  To allow the OpenFaaS function to stay agnostic to the caller it is also possible to register payload transformation templates between the message format and the OpenFaaS function payload.

The listener runs as a standalone application and can be run as a Docker container alongside your OpenFaaS stack, you will also require `gnatsd` to be running.  Information for running nats with OpenFaaS can be found in the Asyncronous calls guide [https://github.com/openfaas/faas/blob/master/guide/asynchronous.md](https://github.com/openfaas/faas/blob/master/guide/asynchronous.md).  Note: this application is not intended to replace asyncronous calls but to allow for implementation of an Event Driven Architecural Pattern where OpenFaaS functions are the unit of work [https://en.wikipedia.org/wiki/Event-driven_architecture](https://en.wikipedia.org/wiki/Event-driven_architecture).

Because the implementation uses Nats.io Subscription queues it is possible to run more than one instance of this application for high availability without suffering duplicate messages.  The message will be delivered to a random instance of faas-nats.

## Configuration
To configure which messages to listen to and functions to call a simple YAML configuration file is used...

```yaml
nats: nats://192.168.1.113:4222
nats_cluster_id: test-cluster
gateway: http://192.168.1.113:8080
statsd: localhost:9125
log_level: DEBUG # TRACE, ERROR, INFO
log_format: text # json
functions:
    # name of the subscription, does not need to correspond to function name
  - name: info
    # function to call upon receipt of message
    function_name: info
    # message to listen to
    message: example.info
 
  - name: echo
    # function to call when an event is received, by default it sends the message
    # payload as received unless a input_template is used
    function_name: echo
    message: example.echo
    # broadcast a message on success of the function, by default it sends the payload
    # as received unless an output template is used
    success_message: example.info.success
    templates:
      # Transform the raw message with a Go template, assumes the payload is json
      input_template: |
        {
          "subject": "{{ .JSON.subject }}"
        }
      # Transform the raw message with a Go template, assumes the payload is json
      output_template: |
          {{printf "%s" .Raw}}
```

## Template functions
### base64encode
Base64 encode []byte

```yaml
input_template: |
{
  "image": "{{ base64encode .Raw }}"
}
```

### base64decode
Base64 decode a string

```yaml
input_template: |
  {{ base64decode .JSON.image }}
```

#### nats
This is a string value with the connection string to your nats server

#### nats_cluster_id
The name of your nats cluster by default this is `test-cluster`

#### gateway
This is a string value corresponding to your OpenFaaS gateway

#### functions
Array of function objects

#### function - name
Name of the subscription, this should be unique and does not need to correspond to the function name

#### function - function_name
Name of the function to call when a message is received

#### function - message
Name of the Nats.io message to listen to

#### function - success_message
Optional string value, upon successful completion of the function call, a message can be broadcast containing the payload returned from the function

#### function - templates
Templates are optional and allow the transformation of the message payload into a function payload and the response from the function into the payload for the
success message.
All templates are in Go template format for more info on Go templates please see: [https://golang.org/pkg/text/template/](https://golang.org/pkg/text/template/)

##### function - templates - output_template
Go format template, before calling the OpenFaaS function the template will be used to process and transform the message

##### function - templates - input_template
Go format template, before publishing the success Nats.io message the template will be used to process and transform the message

## Metrics
Metrics are exported using StatsD to import metrics into Prometheus please use the prometheus StatsD exporter [https://hub.docker.com/r/prom/statsd-exporter/](https://hub.docker.com/r/prom/statsd-exporter/)

## Running the queue
To run the listener you can use the build docker container and provide a configuration file as a volume mount.

```bash
docker run -it \
  -v $(shell pwd)/example_config.yml:/etc/faas-nats/example_config.yml \
  quay.io/nicholasjackson/faas-nats:latest \
  -config /etc/faas-nats/example_config.yml
```

## Testing
There is a simple test harness in ./testharness/main.go which can be used to validate the subscription and transformations.

## TODO
[x] Implement monitoring and metrics with StatsD  
[ ] Handle message wrapping to enable chainging OpenFaaS functions
