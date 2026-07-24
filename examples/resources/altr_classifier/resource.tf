# Copyright (c) ALTR Solutions, Inc.
# SPDX-License-Identifier: Apache-2.0

# Regex classifier
resource "altr_classifier" "ssn" {
  name              = "SSN"
  description       = "Matches US Social Security Numbers"
  pattern           = "\\d{3}-\\d{2}-\\d{4}"
  minimum_threshold = 70
}

# Regex classifier narrowed by a column-name (metadata) filter
resource "altr_classifier" "ssn_named_columns" {
  name              = "SSN_Named_Columns"
  description       = "SSN pattern limited to columns whose name suggests an SSN"
  pattern           = "\\d{3}-\\d{2}-\\d{4}"
  minimum_threshold = 70

  compound_ruleset = jsonencode({
    operator = "AND"
    conditions = [
      {
        target     = "METADATA"
        comparator = "matches"
        pattern    = "(?i)(ssn|social_security|tax_id)"
      }
    ]
  })
}

# Compound ruleset classifier (no regex)
resource "altr_classifier" "unstructured_document" {
  name        = "Unstructured_Document"
  description = "Column dominated by binary documents or images"

  compound_ruleset = jsonencode({
    operator = "OR"
    conditions = [
      {
        target          = "CONTENT_TYPE"
        content_formats = ["PDF", "DOCX", "XLSX", "PNG", "JPEG"]
      }
    ]
  })
}
