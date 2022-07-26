package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

var outputFile string

func init() {
	flag.StringVar(&outputFile, "output", "", "output file for the environment")
}

func main() {
	flag.Parse()

	if outputFile == "" {
		fmt.Fprintln(os.Stderr, "output file must not be empty")
		os.Exit(1)
	}

	cleanedEnviron := []string{}
	for _, v := range os.Environ() {
		subs := strings.SplitN(v, "=", 2)
		if subs[0] != "" {
			cleanedEnviron = append(cleanedEnviron, v)
		}
	}

	bytes, err := json.Marshal(cleanedEnviron)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot marshal environmental variables to JSON: %s", cleanedEnviron)
		os.Exit(1)
	}

	err = ioutil.WriteFile(outputFile, bytes, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot write to output file: '%s'\n", outputFile)
		os.Exit(1)
	}
}
