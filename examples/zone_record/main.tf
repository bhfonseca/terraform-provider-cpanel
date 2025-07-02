provider "cpanel" {
  host      = "cpanel.example.com"
  username  = "your_user"
  api_token = "your_token"
  insecure  = true
}
resource "cpanel_zone_record" "www" {
  zone    = "example.com"
  name    = "www"
  type    = "A"
  address = "192.0.2.1"
  ttl     = 14400
}
