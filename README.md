# terraform-provider-cpanel

Terraform provider to manage DNS zone records via cPanel's API.

## Requirements
- Terraform >= 1.0
- Go >= 1.19
- A cPanel account with API token and proper permissions (ZoneEdit)

## Install
```bash
git clone https://github.com/bhfonseca/terraform-provider-cpanel.git
cd terraform-provider-cpanel
go install
```

Place the compiled binary in:
- Linux/macOS: `~/.terraform.d/plugins/local/custom/cpanel/0.0.1/terraform-provider-cpanel`
- Windows: `%APPDATA%\terraform.d\plugins\local\custom\cpanel\0.0.1\terraform-provider-cpanel.exe`

## Provider Configuration
```hcl
terraform {
  required_providers {
    cpanel = {
      source  = "local/custom/cpanel"
      version = "0.0.1"
    }
  }
}

provider "cpanel" {
  host      = "cpanel.example.com"
  username  = "your_user"
  api_token = "your_token"
  insecure  = true
}
```

## Resources
### `cpanel_zone_record`
Creates a DNS record.

#### Example
```hcl
resource "cpanel_zone_record" "www" {
  zone    = "example.com"
  name    = "www"
  type    = "A"
  address = "192.0.2.1"
  ttl     = 14400
}
```

## License
MIT © Bruno Honorato da Fonseca
