# cPanel Provider

O provedor `cpanel` permite gerenciar registros DNS em servidores cPanel utilizando a API2.

## Autenticação

Este provider utiliza autenticação por token.

```hcl
provider "cpanel" {
  host      = "cpanel.seudominio.com"
  username  = "usuario"
  api_token = "token"
  insecure  = true
}
