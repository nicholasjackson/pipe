output "http" "open_faas" {
  protocol = "http"
  server = "192.168.1.123"
  port = 80
  path = "/"

  tls_config {
   tls_client_key = "key"
   tls_client_cert = "cert"
  }
}
