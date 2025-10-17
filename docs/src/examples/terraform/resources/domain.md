# Addy Domain Resource Example (post full implementation)

```terraform
resource "addy_domain" "managed" {
  domain            = var.domain_name
  description       = var.domain_description
  from_name         = var.domain_from_name
  active            = true
  catch_all         = true
  auto_create_regex = "^prefix"

  lifecycle {
    # If server normalizes regex or description, ignore drift:
    # ignore_changes = [auto_create_regex, description]
  }
}

variable "domain_name" {
  type        = string
  description = "Domain to register in Addy"
}

variable "domain_description" {
  type        = string
  default     = "Managed domain"
  description = "Human friendly description"
}

variable "domain_from_name" {
  type        = string
  default     = "Example Sender"
  description = "Display From name for aliases"
}

output "managed_domain_full" {
  value = {
    id      = addy_domain.managed.id
    domain  = addy_domain.managed.domain
    active  = true
    catch_all = true
  }
}
```
