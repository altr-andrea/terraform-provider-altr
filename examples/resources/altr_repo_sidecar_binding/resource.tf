# Copyright (c) ALTR Solutions, Inc.
# SPDX-License-Identifier: Apache-2.0

resource "altr_sidecar" "example" {
  name     = "example"
  hostname = "example.com"
}

resource "altr_sidecar_listener" "example" {
  sidecar_id         = altr_sidecar.example.id
  port               = 8080
  database_type      = "Oracle"
  advertised_version = "19.0.0.0"
}

resource "altr_repo" "example" {
  name     = "example"
  type     = "Oracle"
  hostname = "example.com"
  port     = 1521
}

resource "altr_repo_sidecar_binding" "example" {
  sidecar_id = altr_sidecar.example.id
  repo_name  = altr_repo.example.name
  port       = altr_sidecar_listener.example.port
}

# MongoDB requires sidecar >= 1.59.0
resource "altr_sidecar_listener" "mongo" {
  sidecar_id         = altr_sidecar.example.id
  port               = 8095
  database_type      = "MongoDB"
  advertised_version = "8.0.0"
}

resource "altr_repo" "mongo" {
  name     = "mongo_db"
  type     = "MongoDB"
  hostname = "mongo-db.example.com"
  port     = 27017
}

resource "altr_repo_sidecar_binding" "mongo" {
  sidecar_id = altr_sidecar.example.id
  repo_name  = altr_repo.mongo.name
  port       = altr_sidecar_listener.mongo.port
}
