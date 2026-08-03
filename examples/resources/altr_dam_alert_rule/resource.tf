# Copyright (c) ALTR Solutions, Inc.
# SPDX-License-Identifier: Apache-2.0

resource "altr_dam_alert_rule" "example" {
  name              = "failed-query-alert"
  description       = "Alert on failed queries against monitored OLTP databases"
  severity          = "high"
  data_source_scope = "oltp"
  rule_type         = "match"
  email_recipients  = ["security@example.com"]

  filter_tree = jsonencode({
    op = "and"
    children = [
      {
        op           = "eq"
        dimension    = "query_status"
        string_value = "failed"
      },
      {
        op         = "in"
        dimension  = "statement_type"
        list_value = ["delete", "drop"]
      },
    ]
  })
}
