package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {
	log.Fatal(http.ListenAndServe(hostPort(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("failed to read body: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		fmt.Println(string(data))
		io.Copy(w, bytes.NewReader(data))
	})))
}

func hostPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return fmt.Sprintf(":%s", port)
}
