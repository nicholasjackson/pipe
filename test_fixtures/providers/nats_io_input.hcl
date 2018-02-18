input "nats_queue" "nats_messages_in" {
  server = "nats://myserver.com"
  cluster_id = "abc123"
  queue = "mymessagequeue"

  auth_basic {
    user = "xxx" // User who has access to the server
    password = "xxx" // Password for user

  }

  auth_mtls {
    tls_client_key = "cakey" // Client key for the streaming server
    tls_client_cert = "cacert" // Client certificate for the streaming server
    tls_client_cacert = "caclient" // Client CA for the streaming server
  }
}
