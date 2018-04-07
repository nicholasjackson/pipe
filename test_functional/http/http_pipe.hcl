input "http" "web_messages_in" {
  protocol = "http"
  server = "localhost"
  port = 18091
  path = "/"
}

output "http" "web_messages_out" {
  protocol = "http"
  server = "localhost"
  port = 18092
  path = "/"
}

pipe "test_web" {
  input = "web_messages_in"

  // do this when a event triggers
  action {
    output = "web_messages_out"
  }
}
