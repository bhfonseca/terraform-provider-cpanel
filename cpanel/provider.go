package cpanel

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			//"cpanel_domain":    resourceDomain(),
			//"cpanel_subdomain": resourceSubdomain(),
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

	err := client.testConnection()
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao cPanel: %s", err)
	}

	return client, nil
}

func (c *cPanelClient) testConnection() error {
	_, err := c.callAPI("SystemInfo", "getversion", map[string]string{})
	return err
}

func (c *cPanelClient) callAPI(module string, function string, params map[string]string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://%s:%d/execute/%s/%s", c.host, c.port, module, function)

	var queryParts []string
	for key, value := range params {
		queryParts = append(queryParts, fmt.Sprintf("%s=%s", key, value))
	}

	if len(queryParts) > 0 {
		url = fmt.Sprintf("%s?%s", url, strings.Join(queryParts, "&"))
	}

	req, err := http.NewRequest("GET", url, nil)
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

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	if status, ok := result["status"].(float64); !ok || status != 1 {
		errorMsg := "erro desconhecido"
		if errors, ok := result["errors"].([]interface{}); ok && len(errors) > 0 {
			errorMsg = fmt.Sprintf("%v", errors[0])
		}
		return nil, fmt.Errorf("erro na API do cPanel: %s", errorMsg)
	}

	return result, nil
}
