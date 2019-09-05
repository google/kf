package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	log.Fatal(http.ListenAndServe(hostPort(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "The current server time is: %s\n", time.Now())
	})))
}

func hostPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return fmt.Sprintf(":%s", port)
}
