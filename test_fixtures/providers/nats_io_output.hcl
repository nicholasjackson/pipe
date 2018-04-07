output "nats_queue" "nats_messages_out" {
  server = "nats://myserver.com"
  cluster_id = "abc123"
  queue = "mymessagequeue"

  auth_basic {
    user = "xxx" // User who has access to the server
    password = "xxx" // Password for user
  }
}
