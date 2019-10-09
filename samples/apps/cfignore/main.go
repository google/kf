package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	http.ListenAndServe(port(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		files := []string{}
		if err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			files = append(files, path)
			return nil
		}); err != nil {
			log.Printf("failed to walkc: %q", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error":%q}`, err)
			return
		}

		if err := json.NewEncoder(w).Encode(files); err != nil {
			log.Printf("failed to encode and write files: %q", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error":%q}`, err)
			return
		}
	}))
}

func port() string {
	if value, ok := os.LookupEnv("PORT"); ok {
		return ":" + value
	}
	return ":8080"
}
