input "nats_queue" "nats_messages_in" {
  server = "nats://${env("nats_server")}:4222"
  cluster_id = "${env("nats_cluster_id")}"
  queue = "messagein"
}

output "nats_queue" "nats_messages_out" {
  server = "nats://${env("nats_server")}:4222"
  cluster_id = "${env("nats_cluster_id")}"
  queue = "messageout"
}

pipe "test_nats" {
  input = "nats_messages_in"

  // do this when a event triggers
  action {
    output = "nats_messages_out"
  }
}
