# Copyright (c) ALTR Solutions, Inc.
# SPDX-License-Identifier: Apache-2.0

resource "altr_classifier_collection" "pii" {
  name        = "pii_collection"
  description = "PII and sensitive data classifiers"
}
