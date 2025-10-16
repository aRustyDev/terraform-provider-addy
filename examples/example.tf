terraform {
  required_providers {
    addy = {
      source = "github.com/aRustyDev/addy"
      version = "0.1.0"
    }
  }
}

provider "addy" {}

resource "addy_domain" "example" {
  name        = "example-domain"
  description = "This is an example domain."
}

data "addy_domain" "example" {}
