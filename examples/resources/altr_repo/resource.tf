# Copyright (c) ALTR Solutions, Inc.
# SPDX-License-Identifier: Apache-2.0

resource "altr_repo" "example" {
  name     = "example"
  type     = "Oracle"
  hostname = "example.com"
  port     = 1521
}

# MongoDB (requires sidecar >= 1.59.0)
resource "altr_repo" "mongo" {
  name     = "mongo_db"
  type     = "MongoDB"
  hostname = "mongo-db.example.com"
  port     = 27017
}
