run "api_token_details" {
  command = plan
  # command = apply

  module {
    source = "../data-sources/addy_api_token_details"
  }

  # variables {
  #   endpoint = run.create_bucket.website_endpoint
  # }

  assert {
    condition     = data.addy_api_token_details.example.status_code == 200
    error_message = "Website responded with HTTP status ${data.http.index.status_code}"
  }
}
