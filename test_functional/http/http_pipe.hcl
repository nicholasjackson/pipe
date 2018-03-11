input "http" "web_messages_in" {
  protocol = "http"
  server = "localhost"
  port = 8091
  path = "/"
}

output "http" "web_messages_out" {
  protocol = "http"
  server = "localhost"
  port = 8092
  path = "/"
}

pipe "test_web" {
  input = "web_messages_in"

  // do this when a event triggers
  action {
    output = "web_messages_out"
  }
}
