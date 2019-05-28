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

package apps

import (
	"fmt"
	"testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/testutil"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
)

func ExampleKfApp() {
	space := NewKfApp()
	// Setup
	space.SetName("nsname")

	// Values
	fmt.Println(space.GetName())

	// Output: nsname
}

func TestKfApp_ToService(t *testing.T) {
	app := NewKfApp()
	app.SetName("foo")
	actual := app.ToService()

	expected := &serving.Service{}
	expected.Name = "foo"

	testutil.AssertEqual(t, "generated service", expected, actual)
}
