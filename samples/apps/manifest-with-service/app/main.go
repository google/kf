package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	envs := map[string]string{}
	for _, e := range os.Environ() {
		x := strings.SplitN(e, "=", 2)
		if len(x) != 2 {
			continue
		}
		envs[x[0]] = x[1]
	}

	log.Fatal(http.ListenAndServe(hostPort(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(envs); err != nil {
			log.Printf("failed to encode envs: %s", err)
		}
	})))
}

func hostPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return fmt.Sprintf(":%s", port)
}
