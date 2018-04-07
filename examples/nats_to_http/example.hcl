input "nats_queue" "nq_in" {
  server = "nats://nats.service.consul:4222"
  cluster_id = "test-cluster"
  queue = "testmessagequeue"
}

output "http" "nq_out" {
  protocol = "http"
  server = "localhost"
  port = 8080
  path = "/message"
}

pipe "accept_nats" {
  input = "nq_in"

  expiration = "1h"

  action {
    output = "nq_out"

    template = <<EOF
      {
        "text": "Hey a picture from selfi drone",
        "image": "{{ .JSON.Data }}"
      }
    EOF

  }
}
