package cpanel

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"io/ioutil"
	"net/http"
	"net/url"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			//"cpanel_domain":      resourceDomain(),
			//"cpanel_subdomain":   resourceSubdomain(),
			"cpanel_zone_record": resourceZoneRecord(),
		},
		DataSourcesMap: map[string]*schema.Resource{},
		Schema: map[string]*schema.Schema{
			"host": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CPANEL_URL", nil),
				Description: "The URL of the cPanel server, e.g., `cpanel.example.com`.",
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CPANEL_USER", nil),
				Description: "The username for the cPanel account.",
			},
			"api_token": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("CPANEL_TOKEN", nil),
				Description: "The API token for the cPanel account.",
			},
			"port": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     2083,
				DefaultFunc: schema.EnvDefaultFunc("CPANEL_PORT", 2083),
				Description: "Porta da API do cPanel (padrão: 2083 para HTTPS)",
			},
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Ignorar verificação SSL",
			},
		},
		ConfigureFunc: providerConfigure,
	}
}

type cPanelClient struct {
	host      string
	username  string
	api_token string
	port      int
	insecure  bool
	client    *http.Client
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	httpClient := &http.Client{}

	if d.Get("insecure").(bool) {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	client := &cPanelClient{
		host:      d.Get("host").(string),
		username:  d.Get("username").(string),
		api_token: d.Get("api_token").(string),
		port:      d.Get("port").(int),
		insecure:  d.Get("insecure").(bool),
		client:    httpClient,
	}

	if err := client.testConnection(); err != nil {
		return nil, fmt.Errorf("erro ao conectar ao cPanel: %s", err)
	}

	return client, nil
}

// testConnection validates that the credentials work by fetching the cPanel version
func (c *cPanelClient) testConnection() error {
	_, err := c.callAPI("SystemInfo", "getversion", nil)
	return err
}

// callAPI is a helper for UAPI (\"/execute\") endpoints
func (c *cPanelClient) callAPI(module, function string, params map[string]string) (map[string]interface{}, error) {
	u := fmt.Sprintf("https://%s:%d/execute/%s/%s", c.host, c.port, module, function)
	if len(params) > 0 {
		q := url.Values{}
		for k, v := range params {
			q.Set(k, v)
		}
		u += "?" + q.Encode()
	}

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "cpanel "+c.username+":"+c.api_token)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res map[string]interface{}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}
	if status, ok := res["status"].(float64); !ok || status != 1 {
		msg := "unknown error"
		if errs, ok := res["errors"].([]interface{}); ok && len(errs) > 0 {
			msg = fmt.Sprintf("%v", errs[0])
		}
		return nil, fmt.Errorf("UAPI error: %s", msg)
	}
	return res, nil
}

// callAPI2 is a helper for legacy cPanel API2 (\"/json-api/cpanel\") endpoints
func (c *cPanelClient) callAPI2(module, function string, params map[string]string) (map[string]interface{}, error) {
	q := url.Values{}
	q.Set("cpanel_jsonapi_user", c.username)
	q.Set("cpanel_jsonapi_apiversion", "2")
	q.Set("cpanel_jsonapi_module", module)
	q.Set("cpanel_jsonapi_func", function)
	for k, v := range params {
		q.Set(k, v)
	}

	u := fmt.Sprintf("https://%s:%d/json-api/cpanel?%s", c.host, c.port, q.Encode())
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "cpanel "+c.username+":"+c.api_token)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res map[string]interface{}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	cp, ok := res["cpanelresult"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}
	if ev, ok := cp["event"].(map[string]interface{}); ok {
		if r, _ := ev["result"].(float64); r == 1 {
			return res, nil
		}
		if reason, ok := ev["reason"].(string); ok {
			return nil, fmt.Errorf("API2 error: %s", reason)
		}
	}
	return nil, fmt.Errorf("API2 unknown failure")
}
