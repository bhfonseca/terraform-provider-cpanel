package cpanel

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceSubdomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceSubdomainCreate,
		Read:   resourceSubdomainRead,
		Delete: resourceSubdomainDelete,
		Schema: map[string]*schema.Schema{
			"subdomain":     {Type: schema.TypeString, Required: true, ForceNew: true},
			"root_domain":   {Type: schema.TypeString, Required: true, ForceNew: true},
			"document_root": {Type: schema.TypeString, Optional: true, ForceNew: true},
		},
	}
}

func resourceSubdomainCreate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*cPanelClient)
	sub := d.Get("subdomain").(string)
	root := d.Get("root_domain").(string)
	dir := d.Get("document_root").(string)
	if dir == "" {
		dir = fmt.Sprintf("public_html/%s.%s", sub, root)
	}

	if _, err := c.callAPI("SubDomain", "addsubdomain", map[string]string{"domain": sub, "rootdomain": root, "dir": dir}); err != nil {
		return err
	}
	d.SetId(fmt.Sprintf("%s.%s", sub, root))
	return resourceSubdomainRead(d, meta)
}

func resourceSubdomainRead(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*cPanelClient)
	res, err := c.callAPI("DomainInfo", "list_domains", nil)
	if err != nil {
		return err
	}
	data, ok := res["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected response format")
	}
	for _, v := range data["sub_domains"].([]interface{}) {
		if v.(string) == d.Id() {
			return nil
		}
	}
	d.SetId("")
	return nil
}

func resourceSubdomainDelete(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*cPanelClient)
	_, err := c.callAPI2("SubDomain", "delsubdomain", map[string]string{"domain": d.Id()})
	if err == nil {
		d.SetId("")
	}
	return err
}
