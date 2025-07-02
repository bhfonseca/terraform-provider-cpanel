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
			"subdomain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"root_domain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"document_root": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
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

	params := map[string]string{
		"domain":     sub,
		"rootdomain": root,
		"dir":        dir,
	}

	if _, err := c.callAPI("SubDomain", "addsubdomain", params); err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s.%s", sub, root))
	return resourceSubdomainRead(d, meta)
}

func resourceSubdomainRead(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*cPanelClient)
	full := d.Id()

	res, err := c.callAPI("DomainInfo", "list_domains", map[string]string{})
	if err != nil {
		return err
	}

	data, ok := res["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("formato de resposta inesperado")
	}

	subs, _ := data["sub_domains"].([]interface{})
	for _, v := range subs {
		if v.(string) == full {
			return nil
		}
	}

	d.SetId("")
	return nil
}

func resourceSubdomainDelete(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*cPanelClient)
	if _, err := c.callAPI("SubDomain", "delsubdomain", map[string]string{"domain": d.Id()}); err != nil {
		return err
	}
	d.SetId("")
	return nil
}
