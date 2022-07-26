package platformoptions

import (
	"encoding/json"
)

type PlatformOptions struct {
	CredhubURI string `json:"credhub-uri"`
}

func Get(jsonPlatformOptions string) (*PlatformOptions, error) {
	if jsonPlatformOptions != "" {
		platformOptions := PlatformOptions{}
		err := json.Unmarshal([]byte(jsonPlatformOptions), &platformOptions)
		if err != nil {
			return nil, err
		}
		return &platformOptions, nil
	}
	return nil, nil
}
