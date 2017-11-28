# OpenFaaS queue listener for Nats.io
This project allows you to listen to Nats.io messages and call OpenFaas functions.  To allow the OpenFaaS function to say agnostic to the caller it is also possible to register payload transformation templates between the message format and the OpenFaaS function payload.

## Configuration
To configure which messages to listen to and functions to call a simple YAML configuration file is used...

```yaml
nats: nats://192.168.1.113:4222
nats_cluster_id: test-cluster
gateway: http://192.168.1.113:8080
functions:
  - name: info
    function_name: info
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

## Running the queue
To run the listener you can use the build docker container and provide a configuration file as a volume mount.

```bash
docker run -it \
  -v $(shell pwd)/example_config.yml:/etc/faas-nats/example_config.yml \
  quay.io/nicholasjackson/faas-nats:latest \
  -config /etc/faas-nats/example_config.yml
```
