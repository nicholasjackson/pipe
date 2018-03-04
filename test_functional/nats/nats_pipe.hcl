input "nats_queue" "nats_messages_in" {
  server = "nats://localhost:4222"
  cluster_id = "abc123"
  queue = "mymessagequeue"
}

output "nats_queue" "nats_messages_out" {
  server = "nats://localhost:4222"
  cluster_id = "abc123"
  queue = "mymessagequeue"
}

pipe "test_nats" {
  input = "nats_messages_in"

  // do this when a event triggers
  action {
    output = "nats_messages_out"
  }
}
