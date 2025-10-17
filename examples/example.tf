terraform {
    required_version = ">= 1.14"
    required_providers {
        addy = {
        source  = "github.com/aRustyDev/addy"
        version = "0.1.0"
        }
    }
}

provider "addy" {
  # Configure the provider with your API key
  # api_key = "your-api-key-here"
  # Or set the ADDY_API_KEY environment variable
  token = var.addy_token
}
