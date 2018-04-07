---
layout: page
homepage: true
---

# Overview

## Installation
Something quick start

## Configuration
Config files for pipe are written using HCL, below is a simple example of a pipe config which would configure an inbound webhook to post a message to nats.io.

```ruby
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
```

### Input providers
Input providers define inputs 

```ruby
input "nats_queue" "nats_messages_in" {
  server = "nats://${env("nats_server")}:4222"
  cluster_id = "${env("nats_cluster_id")}"
  queue = "messagein"
}

```

### Output providers
Output providers define outputs

```ruby
input "nats_queue" "nats_messages_in" {
  server = "nats://${env("nats_server")}:4222"
  cluster_id = "${env("nats_cluster_id")}"
  queue = "messagein"
}
```

### Pipes
Pipes allow you to receive messages from an input provider and send messages to an output provider

```ruby
pipe "test_nats" {
  input = "nats_messages_in"

  // do this when a event triggers
  action {
    output = "nats_messages_out"
  }

  on_success {
    output = "nats_messages_out"
  }

  on_fail {
    output = "nats_messages_out"
  }
}
```
