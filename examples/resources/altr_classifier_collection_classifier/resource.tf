# Copyright (c) ALTR Solutions, Inc.
# SPDX-License-Identifier: Apache-2.0

resource "altr_classifier" "ssn" {
  name              = "SSN"
  description       = "Matches US Social Security Numbers"
  pattern           = "\\d{3}-\\d{2}-\\d{4}"
  minimum_threshold = 70
}

resource "altr_classifier_collection" "pii" {
  name        = "pii_collection"
  description = "PII and sensitive data classifiers"
}

resource "altr_classifier_collection_classifier" "pii_ssn" {
  collection_name = altr_classifier_collection.pii.name
  classifier_name = altr_classifier.ssn.name
}

# Or manage several memberships with for_each over classifiers that
# already exist (e.g. created elsewhere or ALTR-provided).
resource "altr_classifier_collection_classifier" "pii_members" {
  for_each = toset(["Email_Address", "Phone_E164"])

  collection_name = altr_classifier_collection.pii.name
  classifier_name = each.value
}
