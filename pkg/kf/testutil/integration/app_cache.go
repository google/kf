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

package integration

import (
	context "context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/gofrs/flock"
)

const (
	AppCacheEnv = "APP_CACHE"
)

type appCache struct {
	lock *flock.Flock
	path string
}

func newAppCache() *appCache {
	v := os.Getenv(AppCacheEnv)
	if v == "" {
		log.Printf("%q not set, using random file lock", AppCacheEnv)
		f, err := ioutil.TempFile("", "")
		if err != nil {
			log.Fatalf("failed to create lock file: %v", err)
		}
		v = f.Name()
	}

	return &appCache{
		lock: flock.New(v),
		path: v,
	}
}

// Close cleans up any temporary files. If APP_CACHE is set, it will not do
// anything.
func (c *appCache) Close() error {
	if v := os.Getenv(AppCacheEnv); v != "" {
		log.Printf("%q is set, not deletinig file", AppCacheEnv)
		return nil
	}

	if err := c.lock.Close(); err != nil {
		return err
	}
	return os.Remove(c.path)
}

// Load loads the underlying value by a given key from the cache map.
func (c *appCache) Load(ctx context.Context, key string) (string, bool, error) {
	locked, err := c.lock.TryLockContext(ctx, 100*time.Millisecond)
	if err != nil {
		return "", false, err
	} else if !locked {
		return "", false, errors.New("failed to grab lock")
	}
	defer c.lock.Unlock()

	m, err := c.loadMap()
	if err != nil {
		return "", false, err
	}

	existing, ok := m[key]
	return existing, ok, nil
}

// Store sets the value of a given key on the cache map.
func (c *appCache) Store(ctx context.Context, key, value string) error {
	locked, err := c.lock.TryLockContext(ctx, 100*time.Millisecond)
	if err != nil {
		return err
	} else if !locked {
		return errors.New("failed to grab lock")
	}
	defer c.lock.Unlock()

	m, err := c.loadMap()
	if err != nil {
		return err
	}

	// See if the key already exists, if it does move on.
	if _, ok := m[key]; ok {
		return nil
	}
	m[key] = value

	log.Printf("Storing %q (value=%q) into %q", key, value, c.path)
	f, err := os.Create(c.path)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(m); err != nil {
		return fmt.Errorf("failed to encode file: %v", err)
	}
	return nil
}

func (c *appCache) loadMap() (map[string]string, error) {
	m := make(map[string]string)
	f, err := os.Open(c.path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer f.Close()

	// The fine is expected to be encoded in JSON. While this is fairly
	// inefficient, throughput is far from important for this cache.
	if err := json.NewDecoder(f).Decode(&m); err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to decode file: %v", err)
	}

	return m, nil
}
