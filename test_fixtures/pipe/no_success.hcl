pipe "process_image_success" {
  input = "sqs_messages_in"
  // do not process events older than
  expiration = "1hr" 

  // do this when a event triggers
  action {
    output = "api_call"
    template = <<EOF
      {{ print "%s" .Raw }}
    EOF

  }

  // on action fail do this
  on_fail {
    output = "pubsub_outbound"
  }
}
