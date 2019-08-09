package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	http.ListenAndServe(port(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(fmt.Sprintf("got one request from %s", r.Host))
		fmt.Fprintln(w, "hello kf!")
	}))
}

func port() string {
	if value, ok := os.LookupEnv("PORT"); ok {
		return ":" + value
	}
	return ":8080"
}
