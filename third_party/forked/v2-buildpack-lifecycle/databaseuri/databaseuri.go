package databaseuri

import (
	"encoding/json"
	"net/url"
)

type Databaseuri struct {
}

func New() *Databaseuri {
	return &Databaseuri{}
}

func (d *Databaseuri) Credentials(services []byte) ([]string, error) {
	data := map[string][]struct {
		Credentials struct {
			Uri string `json:"uri"`
		} `json:"credentials"`
	}{}
	if err := json.Unmarshal(services, &data); err != nil {
		return nil, err
	}

	var creds []string
	for _, v1 := range data {
		for _, v2 := range v1 {
			if v2.Credentials.Uri != "" {
				creds = append(creds, v2.Credentials.Uri)
			}
		}
	}
	return creds, nil
}

func (d *Databaseuri) Uri(service_uris []string) string {
	schemes := map[string]string{
		"mysql":      "mysql2",
		"mysql2":     "",
		"postgres":   "",
		"postgresql": "postgres",
	}
	for _, service_uri := range service_uris {
		if uri, err := url.Parse(service_uri); err == nil {
			if val, ok := schemes[uri.Scheme]; ok {
				if val != "" {
					uri.Scheme = val
				}
				return uri.String()
			}
		}
	}
	return ""
}
