# Copyright (c) ALTR Solutions, Inc.
# SPDX-License-Identifier: Apache-2.0

resource "altr_sidecar" "example" {
  name     = "example"
  hostname = "example.com"
}

resource "altr_sidecar_listener" "example_8080" {
  sidecar_id         = altr_sidecar.example.id
  port               = 8080
  database_type      = "Oracle"
  advertised_version = "19.0.0.0"
}

resource "altr_sidecar_listener" "example_9000" {
  sidecar_id         = altr_sidecar.example.id
  port               = 9000
  database_type      = "Oracle"
  advertised_version = "19.0.0.0"
}

# MongoDB requires sidecar >= 1.59.0
resource "altr_sidecar_listener" "example_mongo" {
  sidecar_id         = altr_sidecar.example.id
  port               = 8095
  database_type      = "MongoDB"
  advertised_version = "8.0.0"
}
