terraform {
  required_providers {
    addy = {
      source = "aRustyDev/addy"
    }
  }
}

provider "addy" {
  # Configure the provider with your API key
  # api_key = "your-api-key-here"
  # Or set the ADDY_API_KEY environment variable
}

data "addy_api_token_details" "current" {}

output "token_name" {
  value = data.addy_api_token_details.current.name
}

output "token_created_at" {
  value = data.addy_api_token_details.current.created_at
}

output "token_expires_at" {
  value = data.addy_api_token_details.current.expires_at
}