package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var exampleConfig = `
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

    # query string to pass to function
    query_string: abc=123

    # message to listen to
    message: example.info

    # any messages which are older than the expiration time will be ignored and not processed by the system
    # expiration is expressed using Go's duration string format i.e 1000us, 300ms, 1s, 48h, 4d, 1h30m
    expiration: 5s
 
  - name: echo
    # function to call when an event is received, by default it sends the message
    # payload as received unless a input_template is used
    function_name: echo
    message: example.echo
    
    # Transform the raw message with a Go template, assumes the payload is json
    input_template: |
      {
        "subject": "{{ .JSON.subject }}"
      }

    # broadcast n number of messages on success of the function, by default this sends the payload
    # as received unless an output template is used
    success_messages: 
      - name: example.info.success
        # Transform the raw message with a Go template, assumes the payload is json
        output_template: |
          {{printf "%s" .Raw}}
  
      - name: example.detail.success
        output_template: |
          {{printf "%s" .Raw}}
  
    # broadcast n number of messages on failure of the function, by default this sends the 
    # original message payload as received unless an output template is used
    failed_messages:
      - example.info.failed:
        # Transform the raw message with a Go template, assumes the payload is json
        output_template: |
          {{printf "%s" .Raw}}
`

func TestLoadsConfig(t *testing.T) {
	c := Config{}
	err := c.Unmarshal([]byte(exampleConfig))

	assert.Nil(t, err)
}

func TestReturnsErrorOnBadConfig(t *testing.T) {
	c := Config{}
	err := c.Unmarshal([]byte("junk"))

	assert.NotNil(t, err)
}

func TestUnmarshalsFunctions(t *testing.T) {
	c := Config{}
	err := c.Unmarshal([]byte(exampleConfig))

	assert.Nil(t, err)
	assert.Equal(t, len(c.Functions), 2)
	assert.Equal(t, "info", c.Functions[0].Name)
	assert.Equal(t, "info", c.Functions[0].FunctionName)
	assert.Equal(t, "abc=123", c.Functions[0].Query)
	assert.Equal(t, "example.info", c.Functions[0].Message)
	assert.Equal(t, "5s", c.Functions[0].Expiration)
	assert.Contains(t, c.Functions[1].InputTemplate, "\"subject\":")
}

func TestUnmarshalsSuccessMessages(t *testing.T) {
	c := Config{}
	err := c.Unmarshal([]byte(exampleConfig))

	assert.Nil(t, err)
	assert.Equal(t, "example.info.success", c.Functions[1].SuccessMessages[0].Name)
	assert.Contains(t, c.Functions[1].SuccessMessages[0].OutputTemplate, "{{printf ")
}
