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
			"zone":    {Type: schema.TypeString, Required: true, ForceNew: true},
			"name":    {Type: schema.TypeString, Required: true, ForceNew: true, Description: "@ for apex or host"},
			"address": {Type: schema.TypeString, Required: true},
			"ttl":     {Type: schema.TypeInt, Optional: true, Default: 14400},
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

	res, err := c.callAPI("ZoneEdit", "add_zone_record", map[string]string{"domain": zone, "name": host, "type": "A", "address": addr, "ttl": ttl})
	if err != nil {
		return err
	}
	line := int(res["data"].(map[string]interface{})["line"].(float64))
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
	res, err := c.callAPI("ZoneEdit", "fetchzone_records", map[string]string{"domain": zone, "line": strconv.Itoa(line)})
	if err != nil {
		return err
	}
	records := res["data"].([]interface{})
	if len(records) == 0 {
		d.SetId("")
		return nil
	}

	rec := records[0].(map[string]interface{})
	d.Set("address", rec["record"].(string))
	d.Set("ttl", int(rec["ttl"].(float64)))
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
	addr := d.Get("address").(string)
	ttl := strconv.Itoa(d.Get("ttl").(int))

	_, err = c.callAPI("ZoneEdit", "edit_zone_record", map[string]string{"domain": zone, "line": strconv.Itoa(line), "address": addr, "ttl": ttl})
	return err
}

func resourceZoneRecordDelete(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*cPanelClient)
	zone, line, err := parseZoneRecordID(d.Id())
	if err != nil {
		return err
	}
	_, err = c.callAPI("ZoneEdit", "remove_zone_record", map[string]string{"domain": zone, "line": strconv.Itoa(line)})
	if err == nil {
		d.SetId("")
	}
	return err
}
