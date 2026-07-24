// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package classification_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/altrsoftware/terraform-provider-altr/internal/acctest"
)

func TestAccClassifierCollectionResource_basic(t *testing.T) {
	resourceName := "altr_classifier_collection.test"
	name := acctest.RandomWithPrefixUnderscoreMaxLength("coll_test", 24)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckClassifierCollectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClassifierCollectionResourceConfig(name, "Acceptance test collection"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierCollectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "description", "Acceptance test collection"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				// In-place description update via PATCH
				Config: testAccClassifierCollectionResourceConfig(name, "Updated description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassifierCollectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
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

func TestAccCollectionClassifierResource_basic(t *testing.T) {
	resourceName := "altr_classifier_collection_classifier.test"
	prefix := acctest.RandomWithPrefixUnderscoreMaxLength("memb_test", 24)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCollectionClassifierDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionClassifierResourceConfig(prefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionClassifierExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "collection_name", prefix+"_coll"),
					resource.TestCheckResourceAttr(resourceName, "classifier_name", prefix+"_clsf"),
					resource.TestCheckResourceAttr(resourceName, "id", prefix+"_coll:"+prefix+"_clsf"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     prefix + "_coll:" + prefix + "_clsf",
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccClassifierCollectionDataSource_basic(t *testing.T) {
	dataSourceName := "data.altr_classifier_collection.test"
	prefix := acctest.RandomWithPrefixUnderscoreMaxLength("memb_test", 24)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckCollectionClassifierDestroy,
			testAccCheckClassifierCollectionDestroy,
			testAccCheckClassifierDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionClassifierResourceConfig(prefix) + `
data "altr_classifier_collection" "test" {
  name = altr_classifier_collection.test.name

  depends_on = [altr_classifier_collection_classifier.test]
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", prefix+"_coll"),
					resource.TestCheckResourceAttr(dataSourceName, "classifier_names.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "classifier_names.0", prefix+"_clsf"),
				),
			},
		},
	})
}

func testAccClassifierCollectionResourceConfig(name, description string) string {
	return fmt.Sprintf(`
resource "altr_classifier_collection" "test" {
  name        = %[1]q
  description = %[2]q
}
`, name, description)
}

func testAccCollectionClassifierResourceConfig(prefix string) string {
	return fmt.Sprintf(`
resource "altr_classifier" "test" {
  name              = "%[1]s_clsf"
  description       = "Acceptance test classifier"
  pattern           = "\\d{3}-\\d{2}-\\d{4}"
  minimum_threshold = 70
}

resource "altr_classifier_collection" "test" {
  name        = "%[1]s_coll"
  description = "Acceptance test collection"
}

resource "altr_classifier_collection_classifier" "test" {
  collection_name = altr_classifier_collection.test.name
  classifier_name = altr_classifier.test.name
}
`, prefix)
}

func testAccCheckClassifierCollectionExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn, err := testAccClassifierClient()
		if err != nil {
			return fmt.Errorf("failed to create test client: %w", err)
		}

		collection, err := conn.GetClassifierCollection(rs.Primary.Attributes["name"])
		if err != nil {
			return err
		}

		if collection == nil {
			return fmt.Errorf("classifier collection %s does not exist", rs.Primary.Attributes["name"])
		}

		return nil
	}
}

func testAccCheckClassifierCollectionDestroy(s *terraform.State) error {
	conn, err := testAccClassifierClient()
	if err != nil {
		return fmt.Errorf("failed to create test client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "altr_classifier_collection" {
			continue
		}

		collection, err := conn.GetClassifierCollection(rs.Primary.Attributes["name"])
		if err != nil {
			return err
		}

		if collection != nil {
			return fmt.Errorf("classifier collection %s still exists", rs.Primary.Attributes["name"])
		}
	}

	return nil
}

func testAccCheckCollectionClassifierExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn, err := testAccClassifierClient()
		if err != nil {
			return fmt.Errorf("failed to create test client: %w", err)
		}

		classifiers, err := conn.ListClassifierCollectionClassifiers(rs.Primary.Attributes["collection_name"])
		if err != nil {
			return err
		}

		for _, classifier := range classifiers {
			if classifier.Name == rs.Primary.Attributes["classifier_name"] {
				return nil
			}
		}

		return fmt.Errorf("classifier %s is not in collection %s",
			rs.Primary.Attributes["classifier_name"], rs.Primary.Attributes["collection_name"])
	}
}

func testAccCheckCollectionClassifierDestroy(s *terraform.State) error {
	conn, err := testAccClassifierClient()
	if err != nil {
		return fmt.Errorf("failed to create test client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "altr_classifier_collection_classifier" {
			continue
		}

		classifiers, err := conn.ListClassifierCollectionClassifiers(rs.Primary.Attributes["collection_name"])
		if err != nil {
			return err
		}

		for _, classifier := range classifiers {
			if classifier.Name == rs.Primary.Attributes["classifier_name"] {
				return fmt.Errorf("classifier %s is still in collection %s",
					rs.Primary.Attributes["classifier_name"], rs.Primary.Attributes["collection_name"])
			}
		}
	}

	return nil
}
