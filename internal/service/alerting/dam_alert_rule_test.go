// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package alerting_test

import (
	"fmt"
	"testing"

	"github.com/altrsoftware/terraform-provider-altr/internal/acctest"
	"github.com/altrsoftware/terraform-provider-altr/internal/client"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDamAlertRuleResource_basic(t *testing.T) {
	resourceName := "altr_dam_alert_rule.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDamAlertRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDamAlertRuleResourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDamAlertRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "severity", "high"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "data_source_scope", "oltp"),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "match"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDamAlertRuleResource_filterTree(t *testing.T) {
	resourceName := "altr_dam_alert_rule.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDamAlertRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDamAlertRuleResourceConfig_filterTree(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDamAlertRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "email_recipients.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "email_recipients.*", "security@example.com"),
					resource.TestCheckResourceAttrSet(resourceName, "filter_tree"),
				),
			},
			{
				// Re-plan must be empty, or filter_tree isn't round-tripping.
				Config:   testAccDamAlertRuleResourceConfig_filterTree(rName),
				PlanOnly: true,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// jsonencode sorts keys; the API returns its own order.
				ImportStateVerifyIgnore: []string{"filter_tree"},
			},
		},
	})
}

func TestAccDamAlertRuleResource_disappears(t *testing.T) {
	resourceName := "altr_dam_alert_rule.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDamAlertRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDamAlertRuleResourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDamAlertRuleExists(resourceName),
					testAccCheckDamAlertRuleDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDamAlertRuleDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn, err := testClient()
		if err != nil {
			return err
		}

		return conn.DeleteDamAlertRule(rs.Primary.ID)
	}
}

func testAccCheckDamAlertRuleExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is not set")
		}

		conn, err := testClient()
		if err != nil {
			return err
		}

		rule, err := conn.GetDamAlertRule(rs.Primary.ID)
		if err != nil {
			return err
		}

		if rule == nil {
			return fmt.Errorf("DAM alert rule %s not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDamAlertRuleDestroy(s *terraform.State) error {
	conn, err := testClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "altr_dam_alert_rule" {
			continue
		}

		rule, err := conn.GetDamAlertRule(rs.Primary.ID)
		if err != nil {
			return err
		}

		if rule != nil {
			return fmt.Errorf("DAM alert rule %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testClient() (*client.Client, error) {
	conn, err := client.NewClient(
		acctest.TestGetEnv("ALTR_ORG_ID", "test-org"),
		acctest.TestGetEnv("ALTR_API_KEY", "test-key"),
		acctest.TestGetEnv("ALTR_SECRET", "test-secret"),
		acctest.TestGetEnv("ALTR_BASE_URL", ""),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create test client: %w", err)
	}

	return conn, nil
}

func testAccDamAlertRuleResourceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "altr_dam_alert_rule" "test" {
  name              = %[1]q
  severity          = "high"
  data_source_scope = "oltp"
  rule_type         = "match"
}
`, name)
}

func testAccDamAlertRuleResourceConfig_filterTree(name string) string {
	return fmt.Sprintf(`
resource "altr_dam_alert_rule" "test" {
  name              = %[1]q
  description       = "Alert on failed queries"
  severity          = "critical"
  data_source_scope = "oltp"
  rule_type         = "match"
  email_recipients  = ["security@example.com"]

  filter_tree = jsonencode({
    op           = "eq"
    dimension    = "query_status"
    string_value = "failed"
  })
}
`, name)
}
