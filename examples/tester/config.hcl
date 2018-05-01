input "nats_queue" "stream_in" {
  server = "nats://nats.service.consul:4222"
  cluster_id = "test-cluster"
  queue = "image.stream"
}

output "http" "detection_service" {
  protocol = "http"
  server = "drone-detection.service.consul"
  port = 9999
  path = "/detect"
}

output "nats_queue" "drone_new_message" {
  server = "nats://nats.service.consul:4222"
  cluster_id = "test-cluster"
  queue = "image.facedetection"
}

pipe "detect_faces" {
  input = "stream_in"

  action {
    output = "detection_service"
    template = "{{ .JSON.Data }}"
  }

  on_success {
    output = "drone_new_message"
    template = <<EOF
      {
        "faces": {{ tojson(.JSON.Faces) }},
        "bounds": {{ tojson(.JSON.Bounds) }}
      }
EOF

  }
}
