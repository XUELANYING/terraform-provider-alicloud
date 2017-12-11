package alicloud

import (
	"fmt"
	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/slb"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"log"
	"testing"
)

func TestAccAlicloudSlb_basic(t *testing.T) {
	var slb slb.LoadBalancerType

	testCheckAttr := func() resource.TestCheckFunc {
		return func(*terraform.State) error {
			log.Printf("testCheckAttr slb AddressType is: %s", slb.AddressType)
			return nil
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		// module name
		IDRefreshName: "alicloud_slb.bandwidth",

		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSlbDestroy,
		Steps: []resource.TestStep{
			//test internet_charge_type is paybybandwidth
			resource.TestStep{
				Config: testAccSlbBandWidth,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlbExists("alicloud_slb.bandwidth", &slb),
					testCheckAttr(),
					resource.TestCheckResourceAttr(
						"alicloud_slb.bandwidth", "internet_charge_type", "paybybandwidth"),
				),
			},
		},
	})
}

func TestAccAlicloudSlb_traffic(t *testing.T) {
	var slb slb.LoadBalancerType

	testCheckAttr := func() resource.TestCheckFunc {
		return func(*terraform.State) error {
			log.Printf("testCheckAttr slb AddressType is: %s", slb.AddressType)
			return nil
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		// module name
		IDRefreshName: "alicloud_slb.traffic",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckSlbDestroy,
		Steps: []resource.TestStep{
			//test internet_charge_type is paybytraffic
			resource.TestStep{
				Config: testAccSlbTraffic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlbExists("alicloud_slb.traffic", &slb),
					testCheckAttr(),
					resource.TestCheckResourceAttr(
						"alicloud_slb.traffic", "name", "tf_test_slb_classic"),
				),
			},
		},
	})
}

func TestAccAlicloudSlb_vpc(t *testing.T) {
	var slb slb.LoadBalancerType

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		// module name
		IDRefreshName: "alicloud_slb.vpc",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckSlbDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccSlb4Vpc,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlbExists("alicloud_slb.vpc", &slb),
					resource.TestCheckResourceAttr(
						"alicloud_slb.vpc", "name", "tf_test_slb_vpc"),
				),
			},
		},
	})
}

func testAccCheckSlbExists(n string, slb *slb.LoadBalancerType) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SLB ID is set")
		}

		client := testAccProvider.Meta().(*AliyunClient)
		instance, err := client.DescribeLoadBalancerAttribute(rs.Primary.ID)

		if err != nil {
			return err
		}
		if instance == nil {
			return fmt.Errorf("SLB not found")
		}

		*slb = *instance
		return nil
	}
}

func testAccCheckListenersExists(n string, slb *slb.LoadBalancerType, p string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SLB ID is set")
		}

		client := testAccProvider.Meta().(*AliyunClient)
		instance, err := client.DescribeLoadBalancerAttribute(rs.Primary.ID)

		if err != nil {
			return err
		}
		if instance == nil {
			return fmt.Errorf("SLB not found")
		}

		exist := false
		for _, listener := range instance.ListenerPortsAndProtocol.ListenerPortAndProtocol {
			if listener.ListenerProtocol == p {
				exist = true
				break
			}
		}

		if !exist {
			return fmt.Errorf("The %s protocol Listener not found.", p)
		}
		return nil
	}
}

func testAccCheckSlbDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*AliyunClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "alicloud_slb" {
			continue
		}

		// Try to find the Slb
		instance, err := client.DescribeLoadBalancerAttribute(rs.Primary.ID)

		if instance != nil {
			return fmt.Errorf("SLB still exist")
		}

		if err != nil {
			e, _ := err.(*common.Error)
			// Verify the error is what we want
			if e.ErrorResponse.Code != LoadBalancerNotFound {
				return err
			}

		}

	}

	return nil
}

const testAccSlbBandWidth = `
resource "alicloud_slb" "bandwidth" {
  name = "tf_test_slb_bandwidth"
  internet_charge_type = "paybybandwidth"
  bandwidth = 5
  internet = true
}
`

const testAccSlbTraffic = `
resource "alicloud_slb" "traffic" {
  name = "tf_test_slb_classic"
}
`

const testAccSlb4Vpc = `
data "alicloud_zones" "default" {
	"available_resource_creation"= "VSwitch"
}

resource "alicloud_vpc" "foo" {
  name = "tf_test_foo"
  cidr_block = "172.16.0.0/12"
}

resource "alicloud_vswitch" "foo" {
  vpc_id = "${alicloud_vpc.foo.id}"
  cidr_block = "172.16.0.0/21"
  availability_zone = "${data.alicloud_zones.default.zones.0.id}"
}

resource "alicloud_slb" "vpc" {
  name = "tf_test_slb_vpc"
  //internet_charge_type = "paybybandwidth"
  vswitch_id = "${alicloud_vswitch.foo.id}"
}
`
