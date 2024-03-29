# terraform-provider-hatenablog-members

A terraform provider for managing members of the Hatena Blog.

## example

```hcl
terraform {
  required_providers {
    hatenablog-members = {
      source = "hatena/hatenablog-members"
      version = "0.1.0"
    }
  }
}

variable "HATENABLOG_APIKEY" {
  type = string
}
    
provider "hatenablog-members" { 
  username = "hatenablog-tf-test"
  apikey = var.HATENABLOG_APIKEY
  blog_host = "tf-test.hatenablog.com"
}

resource "hatenablog-members_member" "member" {
  username = "hatenablog-tf-test2"
  role = "admin"
}
```
