# Pipe - Event Grid and Message Router

[![Docker Repository on Quay](https://quay.io/repository/nicholasjackson/faas-nats/status "Docker Repository on Quay")](https://quay.io/repository/nicholasjackson/pipe)
[![CircleCI](https://circleci.com/gh/nicholasjackson/pipe.svg?style=svg)](https://circleci.com/gh/nicholasjackson/pipe) 
[![Maintainability](https://api.codeclimate.com/v1/badges/a3c44667f431244a86ae/maintainability)](https://codeclimate.com/github/nicholasjackson/pipe/maintainability)

This project allows you to listen to a variety of message sources and perform an action when a message is received.  The documentation and the project is currently work in progress however curretly supported providers are:
* Nats.io - read and write to nats streaming
* HTTP - receive and send events over http

The project is built around a provider model where plugable elements can be added to the server to allow support for a variety of message sources.

Planned providers:
* Log files - read and write to log files
* SQS - AWS Simple Message Queue
* PubSub - Google pub sub
* Kafka
* And more.

## Configuration
To configure pipes HCL configuration file is used...

```yaml
# Input block, will listen for nats messages on defined queue
input "nats_queue" "nq_in" {
  server = "nats://nats.service.consul:4222"
  cluster_id = "test-cluster"
  queue = "testmessagequeue"
}

# Output block, defines a http output
output "http" "nq_out" {
  protocol = "http"
  server = "localhost"
  port = 8080
  path = "/message"
}

pipe "accept_nats" {
  # Name of the input block
  input = "nq_in"

  # Do not handle messages older than
  expiration = "1h"

  # Action to perform when a new message is received
  action {
    # Name of the output
    output = "nq_out"

    # Transform the initial message
    template = <<EOF
      {
        "text": "Hey a picture from selfi drone",
        "image": "{{ .JSON.Data }}"
      }
    EOF

  }

  # Called when action succeedes
  on_success {
    output = "success"
  }
 
  on_success {
    output = "success"
  }

  # Called when the action fails
  on_fail {
    output = "fail"
  }
}
```

## Template values
### .Raw
Return raw binary data as an array of bytes from the message

### .JSON
If the message type is application/json return an object which allows access to elements
i.e.   
Given:  
```json
{
  "Pets": [
    {"name": "fido"}
  ]
}
```

Then:  
```
  {{ .JSON.Pets[0].name }} // fido
```

Note .JSON does not convert the output to JSON format, writing the direct output of .JSON.Pets would produce a go formatted
object.  To output json see the template function `tojson`.

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
  {{ base64decode .JSON.Image }}
```

### tojson
Convert to valid json
```yaml
input_template: |
  {{ tojson .JSON.Pets }}
```

## Metrics
Metrics are exported using StatsD to import metrics into Prometheus please use the prometheus StatsD exporter [https://hub.docker.com/r/prom/statsd-exporter/](https://hub.docker.com/r/prom/statsd-exporter/)

## Running the queue
To run the listener you can use the build docker container and provide a configuration file as a volume mount.

```bash
docker run -it \
  -v $(shell pwd)/examples:/etc/config \
  quay.io/nicholasjackson/faas-nats:latest \
  -config /etc/config/examples
```

## Testing
There is a simple test harness in ./testharness/main.go which can be used to validate the subscription and transformations.

## TODO
[x] Implement monitoring and metrics with StatsD  
[ ] Finish documentation
[ ] Write more examples
[ ] Finish basic provider implementation
