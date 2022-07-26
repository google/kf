package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

func main() {
	// Copy the launcher and builder to /workspace.
	log.SetFlags(0)

	// Copy the launcher and builder to /workspace.
	if err := copyFile("/workspace/launcher", koDataPath("/launcher")); err != nil {
		log.Fatalf("failed to copy /launcher: %v", err)
	}
	if err := copyFile("/workspace/builder", koDataPath("/builder")); err != nil {
		log.Fatalf("failed to copy /builder: %v", err)
	}
}

func koDataPath(name string) string {
	return filepath.Join(os.Getenv("KO_DATA_PATH"), name)
}

func copyFile(dst, src string) error {
	r, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source: %v", err)
	}
	defer r.Close()

	fi, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat file: %v", err)
	}

	w, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fi.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination: %v", err)
	}
	defer w.Close()

	if _, err := io.Copy(w, r); err != nil {
		return fmt.Errorf("failed to copy to destination: %v", err)
	}

	return nil
}
