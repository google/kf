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

package gentest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/kf/pkg/kf/testutil"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	testclient "k8s.io/client-go/kubernetes/fake"
	cv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func TestAllPredicate(t *testing.T) {
	pass := func(obj *v1.Secret) bool {
		return true
	}

	fail := func(obj *v1.Secret) bool {
		return false
	}

	cases := map[string]struct {
		Children []Predicate
		Expected bool
	}{
		"empty true": {
			Children: []Predicate{},
			Expected: true,
		},
		"pass": {
			Children: []Predicate{pass},
			Expected: true,
		},
		"fail": {
			Children: []Predicate{fail},
			Expected: false,
		},
		"mixed": {
			Children: []Predicate{pass, fail, pass},
			Expected: false,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			pred := AllPredicate(tc.Children...)
			actual := pred(nil)
			testutil.AssertEqual(t, "predicate result", tc.Expected, actual)
		})
	}
}

func ExampleList_Filter() {
	first := v1.Secret{}
	first.Name = "ok"

	second := v1.Secret{}
	second.Name = "name-too-long-to-pass"

	list := List{first, second}

	filtered := list.Filter(func(s *v1.Secret) bool {
		return len(s.Name) < 8
	})

	fmt.Println("Results")
	for _, v := range filtered {
		fmt.Println("-", v.Name)
	}

	// Output: Results
	// - ok
}

func ExampleMutatorList_Apply() {
	mutators := MutatorList{
		func(s *v1.Secret) error {
			s.Name = "Name"
			return nil
		},
		func(s *v1.Secret) error {
			return errors.New("some-error")
		},
	}
	res := v1.Secret{}
	err := mutators.Apply(&res)

	fmt.Println("Error:", err)
	fmt.Println("Mutated name:", res.Name)

	// Output: Error: some-error
	// Mutated name: Name
}

func ExampleLabelSetMutator() {
	out := &v1.Secret{}
	managedAdder := LabelSetMutator(map[string]string{"managed-by": "kf"})

	managedAdder(out)
	fmt.Printf("Labels: %v", out.Labels)

	// Output: Labels: map[managed-by:kf]
}

func ExampleLabelEqualsPredicate() {
	out := &v1.Secret{}
	out.Labels = map[string]string{"managed-by": "not kf"}
	pred := LabelEqualsPredicate("managed-by", "kf")

	fmt.Printf("Not Equal: %v\n", pred(out))

	out.Labels["managed-by"] = "kf"
	fmt.Printf("Equal: %v\n", pred(out))

	// Output: Not Equal: false
	// Equal: true
}

func ExampleLabelsContainsPredicate() {
	out := &v1.Secret{}
	out.Labels = map[string]string{"my-label": ""}

	mylabelpred := LabelsContainsPredicate("my-label")
	missinglabelpred := LabelsContainsPredicate("missing")

	fmt.Printf("Contained: %v\n", mylabelpred(out))
	fmt.Printf("Not Contained: %v\n", missinglabelpred(out))

	// Output: Contained: true
	// Not Contained: false
}

func TestClient_invariant(t *testing.T) {
	// This test validates that the filters and mutators are applied to read and
	// write operations.
	mockK8s := testclient.NewSimpleClientset().CoreV1()

	invalid := &v1.Secret{}
	invalid.Name = "does-not-belong"

	if _, err := mockK8s.Secrets("default").Create(invalid); err != nil {
		t.Fatal(err)
	}

	secretsClient := NewExampleClient(mockK8s)

	t.Run("create", func(t *testing.T) {
		good := &v1.Secret{}
		good.Name = "created-through-client"

		if _, err := secretsClient.Create("default", good); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("list", func(t *testing.T) {
		out, err := secretsClient.List("default")
		testutil.AssertNil(t, "list err", err)

		testutil.AssertEqual(t, "secret count", 1, len(out))
	})

	t.Run("get", func(t *testing.T) {
		_, err := secretsClient.Get("default", "does-not-belong")
		testutil.AssertErrorsEqual(t, errors.New("an object with the name does-not-belong exists, but it doesn't appear to be a OperatorConfig"), err)
	})

	t.Run("transform", func(t *testing.T) {
		err := secretsClient.Transform("default", "created-through-client", func(s *v1.Secret) error {
			s.Labels["mutated"] = "true"
			s.Labels["is-a"] = "try-to-overwrite"

			return nil
		})
		testutil.AssertNil(t, "transform err", err)

		modified, err := secretsClient.Get("default", "created-through-client")
		testutil.AssertNil(t, "get err", err)

		testutil.AssertEqual(t, "mutated label", "true", modified.Labels["mutated"])
		testutil.AssertEqual(t, "is-a label", "OperatorConfig", modified.Labels["is-a"])
	})
}

func TestClient_Delete(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		Namespace string
		Name      string
		Options   []DeleteOption
		ExpectErr error
		Setup     func(mockK8s cv1.SecretsGetter)
	}{
		"secret does not exist": {
			Namespace: "default",
			Name:      "some-secret",
			Options:   []DeleteOption{},
			ExpectErr: errors.New(`couldn't delete the OperatorConfig with the name "some-secret": secrets "some-secret" not found`),
		},
		"secret exists": {
			Namespace: "my-namespace",
			Name:      "some-secret",
			Options:   []DeleteOption{},
			Setup: func(mockK8s cv1.SecretsGetter) {
				secret := &v1.Secret{}
				secret.Name = "some-secret"
				mockK8s.Secrets("my-namespace").Create(secret)
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {

			mockK8s := testclient.NewSimpleClientset().CoreV1()
			if tc.Setup != nil {
				tc.Setup(mockK8s)
			}

			secretsClient := coreClient{
				kclient: mockK8s,
			}

			actualErr := secretsClient.Delete(tc.Namespace, tc.Name, tc.Options...)
			if tc.ExpectErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectErr, actualErr)

				return
			}

			secrets, err := mockK8s.Secrets(tc.Namespace).List(metav1.ListOptions{})
			if err != nil {
				t.Fatal(err)
			}

			for _, s := range secrets.Items {
				if s.Name == tc.Name {
					t.Fatal("The secret wasn't deleted")
				}
			}
		})
	}
}

func TestClient_Upsert(t *testing.T) {
	fakeSecret := func(name string, value string) *v1.Secret {
		s := &v1.Secret{}
		s.Name = name
		s.StringData = map[string]string{"value": value}

		return s
	}
	cases := map[string]struct {
		Namespace   string
		PreExisting []*v1.Secret
		ToUpsert    *v1.Secret
		Merger      Merger
		ExpectErr   error
		Validate    func(t *testing.T, mockK8s cv1.SecretsGetter)
	}{
		"inserts if not found": {
			Namespace: "default",
			ToUpsert:  fakeSecret("foo", "new"),
			Merger:    nil, // should not be called
			Validate: func(t *testing.T, mockK8s cv1.SecretsGetter) {
				secret, err := mockK8s.Secrets("default").Get("foo", metav1.GetOptions{})
				testutil.AssertNil(t, "secrets err", err)
				testutil.AssertEqual(t, "value", "new", secret.StringData["value"])
			},
		},
		"update if found": {
			PreExisting: []*v1.Secret{fakeSecret("foo", "old")},
			Namespace:   "testing",
			ToUpsert:    fakeSecret("foo", "new"),
			Merger:      func(n, o *v1.Secret) *v1.Secret { return n },
			Validate: func(t *testing.T, mockK8s cv1.SecretsGetter) {
				secret, err := mockK8s.Secrets("testing").Get("foo", metav1.GetOptions{})
				testutil.AssertNil(t, "secrets err", err)
				testutil.AssertEqual(t, "value", "new", secret.StringData["value"])
			},
		},
		"calls merge with right order": {
			Namespace:   "default",
			PreExisting: []*v1.Secret{fakeSecret("foo", "old")},
			ToUpsert:    fakeSecret("foo", "new"),
			Merger: func(n, o *v1.Secret) *v1.Secret {
				n.StringData["value"] = n.StringData["value"] + "-" + o.StringData["value"]
				return n
			},
			ExpectErr: nil,
			Validate: func(t *testing.T, mockK8s cv1.SecretsGetter) {
				secret, err := mockK8s.Secrets("default").Get("foo", metav1.GetOptions{})
				testutil.AssertNil(t, "secrets err", err)
				testutil.AssertEqual(t, "value", "new-old", secret.StringData["value"])
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			mockK8s := testclient.NewSimpleClientset().CoreV1()

			secretsClient := NewExampleClient(mockK8s)

			for _, v := range tc.PreExisting {
				_, err := secretsClient.Create(tc.Namespace, v)
				testutil.AssertNil(t, "creating preexisting secrets", err)
			}

			_, actualErr := secretsClient.Upsert(tc.Namespace, tc.ToUpsert, tc.Merger)
			if tc.ExpectErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectErr, actualErr)

				return
			}

			if tc.Validate != nil {
				tc.Validate(t, mockK8s)
			}
		})
	}
}

func ExampleDiffWrapper_noDiff() {
	secret := &v1.Secret{}

	wrapper := DiffWrapper(os.Stdout, func(s *v1.Secret) error {
		// don't mutate the secret
		return nil
	})

	wrapper(secret)

	// Output: No changes
}

func ExampleDiffWrapper_changes() {
	secret := &v1.Secret{}
	secret.Type = "opaque"

	contents := &bytes.Buffer{}
	wrapper := DiffWrapper(contents, func(s *v1.Secret) error {
		s.Type = "docker-creds"
		return nil
	})

	fmt.Println("Error:", wrapper(secret))
	firstLine := strings.Split(contents.String(), "\n")[0]
	fmt.Println("First line:", firstLine)

	// Output: Error: <nil>
	// First line: OperatorConfig Diff (-old +new):
}

func ExampleDiffWrapper_err() {
	secret := &v1.Secret{}

	wrapper := DiffWrapper(os.Stdout, func(s *v1.Secret) error {
		return errors.New("some-error")
	})

	fmt.Println(wrapper(secret))

	// Output: some-error
}

func TestConditionDeleted(t *testing.T) {
	cases := map[string]struct {
		apiErr error

		wantDone bool
		wantErr  error
	}{
		"not found error": {
			apiErr:   apierrors.NewNotFound(schema.GroupResource{}, "my-secret"),
			wantDone: true,
		},
		"nil error": {
			apiErr:   nil,
			wantDone: false,
		},
		"other error": {
			apiErr:   errors.New("some-error"),
			wantDone: true,
			wantErr:  errors.New("some-error"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actualDone, actualErr := ConditionDeleted(nil, tc.apiErr)

			testutil.AssertErrorsEqual(t, tc.wantErr, actualErr)
			testutil.AssertEqual(t, "done", tc.wantDone, actualDone)
		})
	}
}

func ExampleClient_WaitForE_conditionDeleted() {
	mockK8s := testclient.NewSimpleClientset().CoreV1()
	secretsClient := NewExampleClient(mockK8s)

	instance, err := secretsClient.WaitForE(context.Background(), "default", "secret-name", 1*time.Second, ConditionDeleted)
	fmt.Println("Instance:", instance)
	fmt.Println("Error:", err)

	// Output: Instance: nil
	// Error: <nil>
}

func ExampleClient_WaitForE_timeout() {
	mockK8s := testclient.NewSimpleClientset().CoreV1()
	secretsClient := NewExampleClient(mockK8s)

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	called := 0
	instance, err := secretsClient.WaitForE(ctx, "default", "secret-name", 100*time.Millisecond, func(_ *v1.Secret, _ error) (bool, error) {
		called++
		return false, nil
	})
	fmt.Println("Instance:", instance)
	fmt.Println("Error:", err)
	fmt.Println("Called?:", called) // 3 calls, immediately, 100ms later then 100ms after

	// Output: Instance: nil
	// Error: waiting for OperatorConfig timed out
	// Called?: 3
}

func TestWrapPredicate(t *testing.T) {
	cases := map[string]struct {
		apiErr            error
		predicateResponse bool

		wantCall bool
		wantDone bool
		wantErr  error
	}{
		"errors are final": {
			apiErr:   errors.New("some-error"),
			wantDone: true,
			wantErr:  errors.New("some-error"),
		},
		"no error returns response (false)": {
			apiErr:            nil,
			predicateResponse: false,
			wantCall:          true,
			wantDone:          false,
		},
		"no error returns response (true)": {
			apiErr:            nil,
			predicateResponse: true,
			wantCall:          true,
			wantDone:          true,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			called := false
			wrapped := wrapPredicate(func(_ *v1.Secret) bool {
				called = true
				return tc.predicateResponse
			})
			actualDone, actualErr := wrapped(nil, tc.apiErr)

			testutil.AssertErrorsEqual(t, tc.wantErr, actualErr)
			testutil.AssertEqual(t, "done", tc.wantDone, actualDone)
			testutil.AssertEqual(t, "predicate called", tc.wantCall, called)
		})
	}
}
