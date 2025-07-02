package cpanel

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strings"
)

func resourceDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceDomainCreate,
		Read:   resourceDomainRead,
		Delete: resourceDomainDelete,
		Schema: map[string]*schema.Schema{
			"domain":        {Type: schema.TypeString, Required: true, ForceNew: true},
			"subdomain":     {Type: schema.TypeString, Optional: true, Computed: true, ForceNew: true},
			"document_root": {Type: schema.TypeString, Optional: true, ForceNew: true},
		},
	}
}

func resourceDomainCreate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*cPanelClient)
	domain := d.Get("domain").(string)
	sub := d.Get("subdomain").(string)
	if sub == "" {
		sub = strings.Split(domain, ".")[0]
	}
	dir := d.Get("document_root").(string)
	if dir == "" {
		dir = fmt.Sprintf("public_html/%s", domain)
	}

	if _, err := c.callAPI2("AddonDomain", "addaddondomain", map[string]string{"newdomain": domain, "subdomain": sub, "dir": dir}); err != nil {
		return err
	}
	d.SetId(domain)
	d.Set("subdomain", sub)
	return resourceDomainRead(d, meta)
}

func resourceDomainRead(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*cPanelClient)
	res, err := c.callAPI("DomainInfo", "list_domains", nil)
	if err != nil {
		return err
	}
	data, ok := res["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected response format")
	}
	for _, v := range data["addon_domains"].([]interface{}) {
		if v.(string) == d.Id() {
			return nil
		}
	}
	d.SetId("")
	return nil
}

func resourceDomainDelete(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*cPanelClient)
	_, err := c.callAPI2("AddonDomain", "deladdondomain", map[string]string{"domain": d.Id()})
	if err == nil {
		d.SetId("")
	}
	return err
}
