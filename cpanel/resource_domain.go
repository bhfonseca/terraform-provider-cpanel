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
			"domain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"subdomain": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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

func resourceDomainCreate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*cPanelClient)

	domain := d.Get("domain").(string)
	sub := d.Get("subdomain").(string)
	if sub == "" {
		parts := strings.Split(domain, ".")
		if len(parts) < 2 {
			return fmt.Errorf("não foi possível derivar o subdomínio, defina explicitamente 'subdomain'")
		}
		sub = parts[0]
	}

	dir := d.Get("document_root").(string)
	if dir == "" {
		dir = fmt.Sprintf("public_html/%s", domain)
	}

	params := map[string]string{
		"newdomain": domain,
		"subdomain": sub,
		"dir":       dir,
	}

	if _, err := c.callAPI("AddonDomain", "addaddondomain", params); err != nil {
		return err
	}

	d.SetId(domain)
	d.Set("subdomain", sub)
	return resourceDomainRead(d, meta)
}

func resourceDomainRead(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*cPanelClient)
	res, err := c.callAPI("DomainInfo", "list_domains", map[string]string{})
	if err != nil {
		return err
	}

	data, ok := res["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("formato de resposta inesperado")
	}

	addons, _ := data["addon_domains"].([]interface{})
	wanted := d.Id()
	for _, v := range addons {
		if v.(string) == wanted {
			return nil
		}
	}

	// Not found
	d.SetId("")
	return nil
}

func resourceDomainDelete(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*cPanelClient)
	if _, err := c.callAPI("AddonDomain", "deladdondomain", map[string]string{"domain": d.Id()}); err != nil {
		return err
	}
	d.SetId("")
	return nil
}
