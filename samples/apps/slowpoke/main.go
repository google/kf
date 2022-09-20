// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	bindDelay := lookupEnvSeconds("BIND_DELAY", 0)
	healthDelay := lookupEnvSeconds("HEALTH_DELAY", 0) + bindDelay
	healthProbeLatency := lookupEnvSeconds("HEALTH_PROBE_LATENCY", 0)

	log.Println("Bind delay (from startup)", bindDelay)
	log.Println("Health delay (from startup)", healthDelay)
	log.Println("Health latency", healthProbeLatency)

	startTime := time.Now()
	log.Println("Start at", startTime)
	log.Println("Bind will happen at", startTime.Add(bindDelay))
	healthyTime := startTime.Add(healthDelay)
	log.Println("Health will happen at", healthyTime)

	time.Sleep(bindDelay)

	log.Fatal(http.ListenAndServe(hostPort(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Got request, waiting %s\n", healthProbeLatency)
		time.Sleep(healthProbeLatency)

		if time.Now().Before(healthyTime) {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Time before: %q", healthyTime)
			return
		}

		w.WriteHeader(http.StatusOK)
	})))
}

func lookupEnvSeconds(key string, defaultValue int) time.Duration {
	keyStr, ok := os.LookupEnv(key)
	if !ok {
		return time.Duration(defaultValue) * time.Second
	}

	envVal, err := strconv.Atoi(keyStr)
	if err != nil {
		panic(err)
	}
	return time.Duration(envVal) * time.Second
}

func hostPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return fmt.Sprintf(":%s", port)
}
