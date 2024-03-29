package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestBlogMember(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
					resource "hatenablog-members_member" "tf-test2" {
					  username = "hatenablog-tf-test2"
					  role = "admin"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hatenablog-members_member.tf-test2", "username", "hatenablog-tf-test2"),
					resource.TestCheckResourceAttr("hatenablog-members_member.tf-test2", "role", "admin"),
				),
			},
			{
				Config: providerConfig + `
					resource "hatenablog-members_member" "tf-test2" {
					  username = "hatenablog-tf-test2"
					  role = "editor"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hatenablog-members_member.tf-test2", "role", "editor"),
				),
			},
			{
				// cleanup
				Config: providerConfig,
			},
		},
	})
}
