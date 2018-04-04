input "http" "example_in" {
  protocol = "http"
  server = "0.0.0.0"
  port = 9099
  path = "/message"
}

output "nats_queue" "example_out" {
  server = "nats://nats.service.consul:4222"
  cluster_id = "test-cluster"
  queue = "testmessagequeue"
}

pipe "accept_http" {
  input = "example_in"

  expiration = "1h"

  action {
    output = "example_out"
  }
}
