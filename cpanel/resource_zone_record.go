package cpanel

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
	"strings"
)

func resourceZoneRecord() *schema.Resource {
	return &schema.Resource{
		Create: resourceZoneRecordCreate,
		Read:   resourceZoneRecordRead,
		Update: resourceZoneRecordUpdate,
		Delete: resourceZoneRecordDelete,
		Schema: map[string]*schema.Schema{
			"zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "@ para raiz ou hostname sem o domínio",
			},
			"address": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ttl": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  14400,
			},
		},
	}
}

func fqdn(name, zone string) string {
	if name == "@" || name == zone {
		return zone
	}
	if strings.HasSuffix(name, "."+zone) {
		return name
	}
	return name + "." + zone
}

func resourceZoneRecordCreate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*cPanelClient)
	zone := d.Get("zone").(string)
	host := fqdn(d.Get("name").(string), zone)
	addr := d.Get("address").(string)
	ttl := strconv.Itoa(d.Get("ttl").(int))

	res, err := c.callAPI2("ZoneEdit", "add_zone_record", map[string]string{
		"domain":  zone,
		"name":    host,
		"type":    "A",
		"address": addr,
		"ttl":     ttl,
	})
	if err != nil {
		return err
	}

	cp := res["cpanelresult"].(map[string]interface{})
	dataArr := cp["data"].([]interface{})
	if len(dataArr) == 0 {
		return fmt.Errorf("add_zone_record returned empty data")
	}
	line := int(dataArr[0].(map[string]interface{})["line"].(float64))
	d.SetId(fmt.Sprintf("%s:%d", zone, line))
	return resourceZoneRecordRead(d, meta)
}

func parseZoneRecordID(id string) (zone string, line int, err error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid ID")
	}
	l, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, err
	}
	return parts[0], l, nil
}

func resourceZoneRecordRead(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*cPanelClient)
	zone, line, err := parseZoneRecordID(d.Id())
	if err != nil {
		return err
	}
	res, err := c.callAPI2("ZoneEdit", "fetchzone_records", map[string]string{
		"domain": zone,
		"line":   strconv.Itoa(line),
	})
	if err != nil {
		return err
	}
	cp := res["cpanelresult"].(map[string]interface{})
	dataArr := cp["data"].([]interface{})
	if len(dataArr) == 0 {
		d.SetId("")
		return nil
	}
	rec := dataArr[0].(map[string]interface{})
	if addr, ok := rec["address"].(string); ok {
		d.Set("address", addr)
	}
	if ttl, ok := rec["ttl"].(float64); ok {
		d.Set("ttl", int(ttl))
	}
	return nil
}

func resourceZoneRecordUpdate(d *schema.ResourceData, meta interface{}) error {
	if !d.HasChange("address") && !d.HasChange("ttl") {
		return nil
	}
	c := meta.(*cPanelClient)
	zone, line, err := parseZoneRecordID(d.Id())
	if err != nil {
		return err
	}

	params := map[string]string{
		"domain": zone,
		"line":   strconv.Itoa(line),
	}
	if d.HasChange("address") {
		params["address"] = d.Get("address").(string)
	}
	if d.HasChange("ttl") {
		params["ttl"] = strconv.Itoa(d.Get("ttl").(int))
	}

	_, err = c.callAPI2("ZoneEdit", "edit_zone_record", params)
	return err
}

func resourceZoneRecordDelete(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*cPanelClient)
	zone, line, err := parseZoneRecordID(d.Id())
	if err != nil {
		return err
	}
	_, err = c.callAPI2("ZoneEdit", "remove_zone_record", map[string]string{
		"domain": zone,
		"line":   strconv.Itoa(line),
	})
	if err == nil {
		d.SetId("")
	}
	return err
}
