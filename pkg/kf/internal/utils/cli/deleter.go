// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeleteWaiter is the interface for namespaced objects that can be deleted.
type DeleteWaiter interface {
	Delete(namespace, name string, options *metav1.DeleteOptions) error
	WaitForDeletion(ctx context.Context, namespace, name string, interval time.Duration) error
}

func NewDeleter(typeName string, deleter DeleteWaiter) Deleter {
	return Deleter{
		deleter:  deleter,
		typeName: typeName,
		async:    AsyncFlags{},
	}
}

type Deleter struct {
	deleter  DeleteWaiter
	typeName string
	async    AsyncFlags
}

// Add sets up the deleter for the given command
func (d *Deleter) Add(cmd *cobra.Command) {
	d.async.Add(cmd)
}

// Delete removes the object from the server
func (d *Deleter) Delete(w io.Writer, namespace, name string) error {
	foregroundDelete := metav1.DeletePropagationForeground
	if err := d.deleter.Delete(namespace, name, &metav1.DeleteOptions{
		PropagationPolicy: &foregroundDelete,
	}); err != nil {
		return err
	}

	if d.async.IsSynchronous() {
		fmt.Fprintf(w, "Deleting %s %s in space %s...\n", d.typeName, name, namespace)
		if err := d.deleter.WaitForDeletion(context.Background(), namespace, name, 1*time.Second); err != nil {
			return fmt.Errorf("couldn't delete: %s", err)
		}
		fmt.Fprintln(w, "Deleted")
	} else {
		fmt.Fprintf(w, "Deleting %s %s in space %s asynchronously", d.typeName, name, namespace)
	}

	return nil
}
