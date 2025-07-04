package main

import (
	"github.com/bhfonseca/terraform-provider-cpanel/cpanel"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: cpanel.Provider,
	})
}
