# cpanel_zone_record

This resource manages DNS records in a cPanel zone.

## Example Usage

```hcl
resource "cpanel_zone_record" "www" {
  zone    = "example.com"
  name    = "www"
  type    = "A"
  address = "192.0.2.1"
  ttl     = 14400
}
```