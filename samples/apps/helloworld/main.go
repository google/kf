package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func main() {
	http.ListenAndServe(port(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello from %s!\n", getAppName())
	}))
}

func port() string {
	if value, ok := os.LookupEnv("PORT"); ok {
		return ":" + value
	}
	return ":8080"
}

func getAppName() string {
	if appDesc, ok := os.LookupEnv("VCAP_APPLICATION"); ok {
		vcapApplication := struct {
			Name string `json:"application_name"`
		}{}

		if err := json.Unmarshal([]byte(appDesc), &vcapApplication); err == nil {
			return vcapApplication.Name
		}
	}

	return "<undefined>"
}
