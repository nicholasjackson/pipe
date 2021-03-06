pipe "process_image_fail" {
  input = "sqs_messages_in"
  // do not process events older than
  expiration = "1h" 

  // do this when a event triggers
  action {
    output = "api_call"
    template = <<EOF
      {{ print "%s" .Raw }}
EOF

  }

  // on action success do this
  on_success {
    output = "nats_messages_out"
    // send the contents of the first call as the message
    // body to the outbound provider
    template = <<EOF
      {{ print "%s" 0.Raw }}
EOF

  }

  on_success {
    output = "pubsub_outbound"
  }
}
