
output "token_name" {
  value = data.addy_api_token_details.current.name
}

output "token_created_at" {
  value = data.addy_api_token_details.current.created_at
}

output "token_expires_at" {
  value = data.addy_api_token_details.current.expires_at
}
