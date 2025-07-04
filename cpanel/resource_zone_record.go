package cpanel

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
				Description: "@ for root domain, subdomain for others",
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "A",
				ValidateFunc: validation.StringInSlice([]string{
					"A", "AAAA", "CNAME", "TXT", "MX", "SRV", "NS", "PTR",
				}, false),
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

func resourceZoneRecordCreate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*cPanelClient)
	zone := d.Get("zone").(string)
	host := d.Get("name").(string)
	addr := d.Get("address").(string)
	record_type := d.Get("type").(string)
	ttl := strconv.Itoa(d.Get("ttl").(int))

	_, err := c.callAPI2("ZoneEdit", "add_zone_record", map[string]string{
		"domain":  zone,
		"name":    host,
		"type":    record_type,
		"address": addr,
		"ttl":     ttl,
	})
	if err != nil {
		return err
	}

	// Fallback: buscar o line manualmente
	res, err := c.callAPI2("ZoneEdit", "fetchzone_records", map[string]string{
		"domain": zone,
	})
	if err != nil {
		return fmt.Errorf("registro criado, mas falhou ao buscar line: %v", err)
	}
	cp := res["cpanelresult"].(map[string]interface{})
	dataArr := cp["data"].([]interface{})
	for _, item := range dataArr {
		rec := item.(map[string]interface{})
		if rec["name"] == host+"."+zone+"." && rec["type"] == record_type && rec["address"] == addr {
			line := int(rec["line"].(float64))
			d.SetId(fmt.Sprintf("%s:%d", zone, line))
			return resourceZoneRecordRead(d, meta)
		}
	}
	return fmt.Errorf("registro criado mas não foi possível identificar o ID (line)")
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
