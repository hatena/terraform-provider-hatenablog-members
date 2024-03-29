package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	providerConfig = `
variable "HATENABLOG_APIKEY" {
  type = string
}
provider "hatenablog-members" {
  username = "hatenablog-tf-test"
  apikey = var.HATENABLOG_APIKEY
  blog_host = "tf-test.hatenablog.com"
}
`
)

var (
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"hatenablog-members": providerserver.NewProtocol6WithError(New("test")()),
	}
)
