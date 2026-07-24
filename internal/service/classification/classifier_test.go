// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package classification_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/altrsoftware/terraform-provider-altr/internal/acctest"
	"github.com/altrsoftware/terraform-provider-altr/internal/client"
)

func TestAccClassifierResource_pattern(t *testing.T) {
	resourceName := "altr_classifier.test"
	name := acctest.RandomWithPrefixUnderscoreMaxLength("clsf_test", 24)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckClassifierDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClassifierResourceConfig_pattern(name, "Acceptance test classifier", 70),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "pattern", `\d{3}-\d{2}-\d{4}`),
					resource.TestCheckResourceAttr(resourceName, "minimum_threshold", "70"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				// In-place update via PATCH
				Config: testAccClassifierResourceConfig_pattern(name, "Updated description", 80),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "minimum_threshold", "80"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     name,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccClassifierResource_compoundRuleset(t *testing.T) {
	resourceName := "altr_classifier.test"
	name := acctest.RandomWithPrefixUnderscoreMaxLength("clsf_test", 24)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckClassifierDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClassifierResourceConfig_compoundRuleset(name, `["PDF", "PNG"]`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttrSet(resourceName, "compound_ruleset"),
					resource.TestCheckNoResourceAttr(resourceName, "pattern"),
				),
			},
			{
				// In-place ruleset update via PATCH
				Config: testAccClassifierResourceConfig_compoundRuleset(name, `["PDF", "PNG", "JPEG"]`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "compound_ruleset"),
				),
			},
			{
				// Swap ruleset for pattern: exercises RemoveCompoundRuleset
				// (explicit null) in the same PATCH that sets the pattern.
				Config: testAccClassifierResourceConfig_pattern(name, "Swapped to pattern", 70),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "pattern", `\d{3}-\d{2}-\d{4}`),
					resource.TestCheckNoResourceAttr(resourceName, "compound_ruleset"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     name,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccClassifierDataSource_basic(t *testing.T) {
	dataSourceName := "data.altr_classifier.test"
	name := acctest.RandomWithPrefixUnderscoreMaxLength("clsf_test", 24)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckClassifierDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClassifierResourceConfig_pattern(name, "Acceptance test classifier", 70) + `
data "altr_classifier" "test" {
  name = altr_classifier.test.name
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", name),
					resource.TestCheckResourceAttr(dataSourceName, "pattern", `\d{3}-\d{2}-\d{4}`),
					resource.TestCheckResourceAttr(dataSourceName, "minimum_threshold", "70"),
				),
			},
		},
	})
}

func testAccClassifierResourceConfig_pattern(name, description string, minimumThreshold int) string {
	return fmt.Sprintf(`
resource "altr_classifier" "test" {
  name              = %[1]q
  description       = %[2]q
  pattern           = "\\d{3}-\\d{2}-\\d{4}"
  minimum_threshold = %[3]d
}
`, name, description, minimumThreshold)
}

func testAccClassifierResourceConfig_compoundRuleset(name, contentFormats string) string {
	return fmt.Sprintf(`
resource "altr_classifier" "test" {
  name        = %[1]q
  description = "Acceptance test compound classifier"

  compound_ruleset = jsonencode({
    operator = "OR"
    conditions = [
      {
        target          = "CONTENT_TYPE"
        content_formats = %[2]s
      }
    ]
  })
}
`, name, contentFormats)
}

func testAccClassifierClient() (*client.Client, error) {
	return client.NewClient(
		acctest.TestGetEnv("ALTR_ORG_ID", "test-org"),
		acctest.TestGetEnv("ALTR_API_KEY", "test-key"),
		acctest.TestGetEnv("ALTR_SECRET", "test-secret"),
		acctest.TestGetEnv("ALTR_BASE_URL", ""),
	)
}

func testAccCheckClassifierExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn, err := testAccClassifierClient()
		if err != nil {
			return fmt.Errorf("failed to create test client: %w", err)
		}

		classifier, err := conn.GetClassifier(rs.Primary.Attributes["name"])
		if err != nil {
			return err
		}

		if classifier == nil {
			return fmt.Errorf("classifier %s does not exist", rs.Primary.Attributes["name"])
		}

		return nil
	}
}

func testAccCheckClassifierDestroy(s *terraform.State) error {
	conn, err := testAccClassifierClient()
	if err != nil {
		return fmt.Errorf("failed to create test client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "altr_classifier" {
			continue
		}

		classifier, err := conn.GetClassifier(rs.Primary.Attributes["name"])
		if err != nil {
			return err
		}

		if classifier != nil {
			return fmt.Errorf("classifier %s still exists", rs.Primary.Attributes["name"])
		}
	}

	return nil
}
