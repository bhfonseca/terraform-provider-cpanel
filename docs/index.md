# Cpanel Provider
The `cpanel` provider allows managing DNS records on cPanel servers using the API2.

## Example Usage

```hcl
provider "cpanel" {
  host      = "cpanel.example.com"
  username  = "user"
  api_token = "token"
  port      = 2083 #Optional, default is 2083
  insecure  = true
}
```

## Argunents Reference

The following arguments are supported:

* `host` - (Required) The hostname of the cPanel server.
* `username` - (Required) The username for cPanel.
* `api_token` - (Required) The API token for cPanel.
* `port` - (Optional) The port to connect to the cPanel server. Default is 2083.
* `insecure` - (Optional) If true, allows insecure connections. Default is false.