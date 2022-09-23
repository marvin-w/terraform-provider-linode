package databasemongodb_test

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/linode/linodego"
	"github.com/linode/terraform-provider-linode/linode/acceptance"
	"github.com/linode/terraform-provider-linode/linode/databasemongodb/tmpl"
	"github.com/linode/terraform-provider-linode/linode/helper"
)

var (
	engineVersion string
	testRegion    string
)

func init() {
	resource.AddTestSweepers("linode_database_mongodb", &resource.Sweeper{
		Name: "linode_database_mongodb",
		F:    sweep,
	})

	// TODO: Uncomment when MongoDB is re-enabled
	//client, err := acceptance.GetClientForSweepers()
	//if err != nil {
	//	log.Fatalf("failed to get client: %s", err)
	//}
	//
	//v, err := helper.ResolveValidDBEngine(context.Background(), *client, "mongodb")
	//if err != nil {
	//	log.Fatalf("failde to get db engine version: %s", err)
	//}
	//
	//engineVersion = v.ID

	region, err := acceptance.GetRandomRegionWithCaps([]string{"Managed Databases"})
	if err != nil {
		log.Fatal(err)
	}

	testRegion = region
}

func sweep(prefix string) error {
	client, err := acceptance.GetClientForSweepers()
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}

	listOpts := acceptance.SweeperListOptions(prefix, "label")

	dbs, err := client.ListMongoDatabases(context.Background(), listOpts)
	if err != nil {
		return fmt.Errorf("error getting mongo databases: %s", err)
	}
	for _, db := range dbs {
		if !acceptance.ShouldSweep(prefix, db.Label) {
			continue
		}
		err := client.DeleteMongoDatabase(context.Background(), db.ID)
		if err != nil {
			return fmt.Errorf("error destroying %s during sweep: %s", db.Label, err)
		}
	}

	return nil
}

func TestAccResourceDatabaseMongo_basic(t *testing.T) {
	t.Skip()
	t.Parallel()

	resName := "linode_database_mongodb.foobar"
	dbName := acctest.RandomWithPrefix("tf_test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.TestAccProviders,
		CheckDestroy: checkDestroy,
		Steps: []resource.TestStep{
			{
				Config: tmpl.Basic(t, dbName, engineVersion, testRegion),
				Check: resource.ComposeTestCheckFunc(
					acceptance.CheckMongoDatabaseExists(resName, nil),
					resource.TestCheckResourceAttr(resName, "engine_id", engineVersion),
					resource.TestCheckResourceAttr(resName, "label", dbName),
					resource.TestCheckResourceAttr(resName, "region", testRegion),
					resource.TestCheckResourceAttr(resName, "type", "g6-nanode-1"),

					resource.TestCheckResourceAttr(resName, "allow_list.#", "0"),
					resource.TestCheckResourceAttr(resName, "cluster_size", "1"),
					resource.TestCheckResourceAttr(resName, "encrypted", "false"),
					resource.TestCheckResourceAttr(resName, "storage_engine", "wiredtiger"),
					resource.TestCheckResourceAttr(resName, "compression_type", "none"),
					resource.TestCheckResourceAttr(resName, "ssl_connection", "false"),

					resource.TestCheckResourceAttrSet(resName, "created"),
					resource.TestCheckResourceAttrSet(resName, "host_primary"),
					resource.TestCheckResourceAttrSet(resName, "root_password"),
					resource.TestCheckResourceAttr(resName, "status", "active"),
					resource.TestCheckResourceAttrSet(resName, "updated"),
					resource.TestCheckResourceAttrSet(resName, "root_password"),

					resource.TestCheckResourceAttr(resName, "engine", strings.Split(engineVersion, "/")[0]),
					resource.TestCheckResourceAttr(resName, "version", strings.Split(engineVersion, "/")[1]),
				),
			},
			{
				ResourceName:      resName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceDatabaseMongo_complex(t *testing.T) {
	t.Skip()
	t.Parallel()

	resName := "linode_database_mongodb.foobar"
	dbName := acctest.RandomWithPrefix("tf_test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.TestAccProviders,
		CheckDestroy: checkDestroy,
		Steps: []resource.TestStep{
			{
				Config: tmpl.Complex(t, tmpl.TemplateData{
					Engine:          engineVersion,
					Label:           dbName,
					AllowedIP:       "0.0.0.0/0",
					ClusterSize:     3,
					Encrypted:       true,
					CompressionType: "zlib",
					StorageEngine:   "wiredtiger",
					SSLConnection:   true,
					Region:          testRegion,
				}),
				Check: resource.ComposeTestCheckFunc(
					acceptance.CheckMongoDatabaseExists(resName, nil),
					resource.TestCheckResourceAttr(resName, "engine_id", engineVersion),
					resource.TestCheckResourceAttr(resName, "label", dbName),
					resource.TestCheckResourceAttr(resName, "region", testRegion),
					resource.TestCheckResourceAttr(resName, "type", "g6-nanode-1"),

					resource.TestCheckResourceAttr(resName, "allow_list.#", "1"),
					resource.TestCheckResourceAttr(resName, "allow_list.0", "0.0.0.0/0"),

					resource.TestCheckResourceAttr(resName, "cluster_size", "3"),
					resource.TestCheckResourceAttr(resName, "encrypted", "true"),
					resource.TestCheckResourceAttr(resName, "storage_engine", "wiredtiger"),
					resource.TestCheckResourceAttr(resName, "compression_type", "zlib"),
					resource.TestCheckResourceAttr(resName, "ssl_connection", "true"),

					resource.TestCheckResourceAttr(resName, "updates.#", "1"),
					resource.TestCheckResourceAttr(resName, "updates.0.day_of_week", "saturday"),
					resource.TestCheckResourceAttr(resName, "updates.0.duration", "1"),
					resource.TestCheckResourceAttr(resName, "updates.0.frequency", "monthly"),
					resource.TestCheckResourceAttr(resName, "updates.0.hour_of_day", "22"),
					resource.TestCheckResourceAttr(resName, "updates.0.week_of_month", "2"),

					resource.TestCheckResourceAttrSet(resName, "ca_cert"),
					resource.TestCheckResourceAttrSet(resName, "created"),
					resource.TestCheckResourceAttrSet(resName, "host_primary"),
					resource.TestCheckResourceAttrSet(resName, "root_password"),
					resource.TestCheckResourceAttr(resName, "status", "active"),
					resource.TestCheckResourceAttrSet(resName, "updated"),
					resource.TestCheckResourceAttrSet(resName, "root_password"),

					resource.TestCheckResourceAttr(resName, "engine", strings.Split(engineVersion, "/")[0]),
					resource.TestCheckResourceAttr(resName, "version", strings.Split(engineVersion, "/")[1]),
				),
			},
			{
				Config: tmpl.ComplexUpdates(t, tmpl.TemplateData{
					Engine:          engineVersion,
					Label:           dbName + "updated",
					AllowedIP:       "192.0.2.1/32",
					ClusterSize:     3,
					Encrypted:       true,
					CompressionType: "zlib",
					StorageEngine:   "wiredtiger",
					SSLConnection:   true,
					Region:          testRegion,
				}),
				Check: resource.ComposeTestCheckFunc(
					acceptance.CheckMongoDatabaseExists(resName, nil),
					resource.TestCheckResourceAttr(resName, "engine_id", engineVersion),
					resource.TestCheckResourceAttr(resName, "label", dbName+"updated"),
					resource.TestCheckResourceAttr(resName, "region", testRegion),
					resource.TestCheckResourceAttr(resName, "type", "g6-nanode-1"),

					resource.TestCheckResourceAttr(resName, "allow_list.#", "1"),
					resource.TestCheckResourceAttr(resName, "allow_list.0", "192.0.2.1/32"),

					resource.TestCheckResourceAttr(resName, "cluster_size", "3"),
					resource.TestCheckResourceAttr(resName, "encrypted", "true"),
					resource.TestCheckResourceAttr(resName, "storage_engine", "wiredtiger"),
					resource.TestCheckResourceAttr(resName, "compression_type", "zlib"),
					resource.TestCheckResourceAttr(resName, "ssl_connection", "true"),

					resource.TestCheckResourceAttr(resName, "updates.#", "1"),
					resource.TestCheckResourceAttr(resName, "updates.0.day_of_week", "wednesday"),
					resource.TestCheckResourceAttr(resName, "updates.0.duration", "3"),
					resource.TestCheckResourceAttr(resName, "updates.0.frequency", "weekly"),
					resource.TestCheckResourceAttr(resName, "updates.0.hour_of_day", "13"),

					resource.TestCheckResourceAttrSet(resName, "ca_cert"),
					resource.TestCheckResourceAttrSet(resName, "created"),
					resource.TestCheckResourceAttrSet(resName, "host_primary"),
					resource.TestCheckResourceAttrSet(resName, "root_password"),
					resource.TestCheckResourceAttr(resName, "status", "active"),
					resource.TestCheckResourceAttrSet(resName, "updated"),
					resource.TestCheckResourceAttrSet(resName, "root_password"),

					resource.TestCheckResourceAttr(resName, "engine", strings.Split(engineVersion, "/")[0]),
					resource.TestCheckResourceAttr(resName, "version", strings.Split(engineVersion, "/")[1]),
				),
			},
			{
				ResourceName:      resName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func checkDestroy(s *terraform.State) error {
	client := acceptance.TestAccProvider.Meta().(*helper.ProviderMeta).Client
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "linode_database_mongodb" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error parsing %v to int", rs.Primary.ID)
		}
		if id == 0 {
			return fmt.Errorf("Would have considered %v as %d", rs.Primary.ID, id)
		}

		_, err = client.GetMongoDatabase(context.Background(), id)

		if err == nil {
			return fmt.Errorf("mongo database with id %d still exists", id)
		}

		if apiErr, ok := err.(*linodego.Error); ok && apiErr.Code != 404 {
			return fmt.Errorf("error requesting mongo database with id %d", id)
		}
	}

	return nil
}