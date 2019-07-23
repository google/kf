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

package spaces

import (
	"fmt"
	"strings"
	"testing"

	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/testutil"
)

func ExampleKfSpace() {
	space := NewKfSpace()
	// Setup
	space.SetName("nsname")
	space.SetContainerRegistry("gcr.io/my-registry")

	// Values
	fmt.Println("Name:", space.GetName())
	fmt.Println("Registry:", space.GetContainerRegistry())

	// Output: Name: nsname
	// Registry: gcr.io/my-registry
}

func TestKfSpace_ToSpace(t *testing.T) {
	space := NewKfSpace()
	space.SetName("foo")
	actual := space.ToSpace()

	expected := &v1alpha1.Space{}
	expected.Name = "foo"

	testutil.AssertEqual(t, "generated ns", expected, actual)
}

func ExampleKfSpace_AppendDomain() {
	space := NewKfSpace()
	// Setup
	space.AppendDomains(v1alpha1.SpaceDomain{Domain: "example.com"})
	space.AppendDomains(v1alpha1.SpaceDomain{Domain: "other-example.com"})

	// Values
	var domainNames []string
	for _, domain := range space.Spec.Execution.Domains {
		domainNames = append(domainNames, domain.Domain)
	}
	fmt.Println("Domains:", strings.Join(domainNames, ", "))

	// Output: Domains: example.com, other-example.com
}
