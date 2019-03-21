package secrets

import (
	"errors"
	"sort"
	"testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/testutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	testclient "k8s.io/client-go/kubernetes/fake"
)

func TestClient_Create(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		Name             string
		Options          []CreateOption
		CreateErr        error
		ExpectErr        error
		ExpectNamespace  string
		ExpectStringData map[string]string
		ExpectData       map[string][]byte
		ExpectLabels     map[string]string
	}{
		"use default namespace by default": {
			Name: "broker-secret",
			Options: []CreateOption{
				WithCreateStringData(map[string]string{"username": "user", "password": "pass"}),
			},
			ExpectNamespace:  "default",
			ExpectStringData: map[string]string{"username": "user", "password": "pass"},
		},
		"basicauth": {
			Name: "broker-secret",
			Options: []CreateOption{
				WithCreateNamespace("supersecret"),
				WithCreateStringData(map[string]string{"username": "user", "password": "pass"}),
			},
			ExpectNamespace:  "supersecret",
			ExpectStringData: map[string]string{"username": "user", "password": "pass"},
		},
		"error": {
			Name: "broker-secret",
			Options: []CreateOption{
				WithCreateNamespace("default"),
				WithCreateStringData(map[string]string{"username": "user", "password": "pass"}),
			},
			CreateErr: errors.New("some-error"),
			ExpectErr: errors.New("some-error"),
		},
		"non-string data": {
			Name: "broker-secret",
			Options: []CreateOption{
				WithCreateNamespace("default"),
				WithCreateData(map[string][]byte{"username": []byte("user"), "password": []byte("pass")}),
			},
			ExpectData: map[string][]byte{"username": []byte("user"), "password": []byte("pass")},
		},
		"labels": {
			Name: "broker-secret",
			Options: []CreateOption{
				WithCreateLabels(map[string]string{"key": "value"}),
			},
			ExpectLabels: map[string]string{"key": "value"},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {

			mockK8s := testclient.NewSimpleClientset()
			k8sFactory := func() (kubernetes.Interface, error) {
				return mockK8s, tc.CreateErr
			}

			secretsClient := NewClient(k8sFactory)

			actualErr := secretsClient.Create(tc.Name, tc.Options...)
			if tc.ExpectErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectErr, actualErr)

				return
			}

			secret, err := mockK8s.CoreV1().Secrets(tc.ExpectNamespace).Get(tc.Name, v1.GetOptions{})
			if err != nil {
				t.Fatal(err)
			}

			testutil.AssertEqual(t, "StringData", tc.ExpectStringData, secret.StringData)
			testutil.AssertEqual(t, "Data", tc.ExpectData, secret.Data)
			testutil.AssertEqual(t, "labels", tc.ExpectLabels, secret.Labels)
		})
	}
}

func TestClient_Get(t *testing.T) {
	t.Parallel()

	dummySecret := map[string][]byte{
		"foo": []byte("foo"),
	}

	cases := map[string]struct {
		Name          string
		Options       []GetOption
		FactoryErr    error
		ExpectErr     error
		ExpectSecrets map[string][]byte
		Setup         func(mockK8s kubernetes.Interface)
	}{
		"uses default namespace": {
			Name:          "some-secret",
			Options:       []GetOption{},
			ExpectSecrets: dummySecret,
			Setup: func(mockK8s kubernetes.Interface) {
				secret := &corev1.Secret{Data: dummySecret}
				secret.Name = "some-secret"
				mockK8s.CoreV1().Secrets("default").Create(secret)
			},
		},
		"custom namespace": {
			Name: "some-secret",
			Options: []GetOption{
				WithGetNamespace("custom-namespace"),
			},
			ExpectSecrets: dummySecret,
			Setup: func(mockK8s kubernetes.Interface) {
				secret := &corev1.Secret{Data: dummySecret}
				secret.Name = "some-secret"
				mockK8s.CoreV1().Secrets("custom-namespace").Create(secret)
			},
		},
		"secret does not exist": {
			Name:      "some-secret",
			Options:   []GetOption{},
			ExpectErr: errors.New(`secrets "some-secret" not found`),
		},
		"client creation error": {
			Name:       "some-secret",
			Options:    []GetOption{},
			FactoryErr: errors.New(`some-error`),
			ExpectErr:  errors.New(`some-error`),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {

			mockK8s := testclient.NewSimpleClientset()
			k8sFactory := func() (kubernetes.Interface, error) {
				return mockK8s, tc.FactoryErr
			}

			if tc.Setup != nil {
				tc.Setup(mockK8s)
			}

			secretsClient := NewClient(k8sFactory)
			actualSecrets, actualErr := secretsClient.Get(tc.Name, tc.Options...)
			if tc.ExpectErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectErr, actualErr)

				return
			}

			testutil.AssertEqual(t, "secrets", tc.ExpectSecrets, actualSecrets.Data)
		})
	}
}

func TestClient_Delete(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		Name      string
		Options   []DeleteOption
		DeleteErr error
		ExpectErr error
		Setup     func(mockK8s kubernetes.Interface)
	}{
		"secret does not exist": {
			Name: "some-secret",
			Options: []DeleteOption{
				WithDeleteNamespace("default"),
			},
			ExpectErr: errors.New(`secrets "some-secret" not found`),
		},
		"client creation error": {
			Name: "some-secret",
			Options: []DeleteOption{
				WithDeleteNamespace("default"),
			},
			DeleteErr: errors.New(`some-error`),
			ExpectErr: errors.New(`some-error`),
		},
		"secret exists": {
			Name: "some-secret",
			Options: []DeleteOption{
				WithDeleteNamespace("my-namespace"),
			},
			Setup: func(mockK8s kubernetes.Interface) {
				secret := &corev1.Secret{}
				secret.Name = "some-secret"
				mockK8s.CoreV1().Secrets("my-namespace").Create(secret)
			},
		},
		"uses default namespace": {
			Name:    "some-secret",
			Options: []DeleteOption{},
			Setup: func(mockK8s kubernetes.Interface) {
				secret := &corev1.Secret{}
				secret.Name = "some-secret"
				mockK8s.CoreV1().Secrets("default").Create(secret)
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {

			mockK8s := testclient.NewSimpleClientset()
			k8sFactory := func() (kubernetes.Interface, error) {
				return mockK8s, tc.DeleteErr
			}

			if tc.Setup != nil {
				tc.Setup(mockK8s)
			}

			secretsClient := NewClient(k8sFactory)
			actualErr := secretsClient.Delete(tc.Name, tc.Options...)
			if tc.ExpectErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectErr, actualErr)

				return
			}

			secrets, err := mockK8s.CoreV1().Secrets(DeleteOptions(tc.Options).Namespace()).List(v1.ListOptions{})
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

func createDummySecret(name, ns string, labels map[string]string) func(kubernetes.Interface) {
	return func(mockK8s kubernetes.Interface) {
		secret := &corev1.Secret{}
		secret.Name = name
		secret.Labels = labels
		mockK8s.CoreV1().Secrets(ns).Create(secret)
	}
}

func TestClient_AddLabels(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		Name            string
		Labels          map[string]string
		Options         []AddLabelsOption
		FactoryError    error
		ExpectErr       error
		ExpectNamespace string
		ExpectLabels    map[string]string
		Setup           func(mockK8s kubernetes.Interface)
	}{
		"secret does not exist": {
			Name:            "some-secret",
			Options:         []AddLabelsOption{},
			ExpectNamespace: "default",
			ExpectErr:       errors.New(`secrets "some-secret" not found`),
		},
		"client creation error": {
			Name:         "some-secret",
			Options:      []AddLabelsOption{},
			FactoryError: errors.New(`some-error`),
			ExpectErr:    errors.New(`some-error`),
		},
		"custom values": {
			Name: "some-secret",
			Options: []AddLabelsOption{
				WithAddLabelsNamespace("my-namespace"),
			},
			Labels:          map[string]string{"key1": "val1"},
			Setup:           createDummySecret("some-secret", "my-namespace", nil),
			ExpectNamespace: "my-namespace",
			ExpectLabels:    map[string]string{"key1": "val1"},
		},
		"default values": {
			Name:            "some-secret",
			Options:         []AddLabelsOption{},
			Labels:          nil,
			Setup:           createDummySecret("some-secret", "default", nil),
			ExpectNamespace: "default",
			ExpectLabels:    map[string]string{},
		},
		"adds to existing": {
			Name:            "some-secret",
			Options:         []AddLabelsOption{},
			Labels:          map[string]string{"addk": "addv"},
			Setup:           createDummySecret("some-secret", "default", map[string]string{"origk": "origv"}),
			ExpectNamespace: "default",
			ExpectLabels:    map[string]string{"origk": "origv", "addk": "addv"},
		},
		"clobbers existing": {
			Name:            "some-secret",
			Options:         []AddLabelsOption{},
			Labels:          map[string]string{"k": "new"},
			Setup:           createDummySecret("some-secret", "default", map[string]string{"k": "old"}),
			ExpectNamespace: "default",
			ExpectLabels:    map[string]string{"k": "new"},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {

			mockK8s := testclient.NewSimpleClientset()
			k8sFactory := func() (kubernetes.Interface, error) {
				return mockK8s, tc.FactoryError
			}

			if tc.Setup != nil {
				tc.Setup(mockK8s)
			}

			secretsClient := NewClient(k8sFactory)
			actualErr := secretsClient.AddLabels(tc.Name, tc.Labels, tc.Options...)
			if tc.ExpectErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectErr, actualErr)

				return
			}

			secret, err := mockK8s.CoreV1().Secrets(tc.ExpectNamespace).Get(tc.Name, v1.GetOptions{})
			if err != nil {
				t.Fatal(err)
			}

			testutil.AssertEqual(t, "labels", tc.ExpectLabels, secret.Labels)
		})
	}
}

func TestClient_List(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		Options         []ListOption
		FactoryError    error
		ExpectErr       error
		ExpectNamespace string
		ExpectSecrets   []string
		Setup           func(mockK8s kubernetes.Interface)
	}{
		"client creation error": {
			Options:      []ListOption{},
			FactoryError: errors.New(`some-error`),
			ExpectErr:    errors.New(`some-error`),
		},
		"no secrets": {
			Options:      []ListOption{},
			FactoryError: errors.New(`some-error`),
			ExpectErr:    errors.New(`some-error`),
		},
		"default options": {
			Options:       []ListOption{},
			Setup:         createDummySecret("secret1", "default", nil),
			ExpectSecrets: []string{"secret1"},
		},
		"custom namespace": {
			Options:       []ListOption{WithListNamespace("custom-ns")},
			Setup:         createDummySecret("secret1", "custom-ns", nil),
			ExpectSecrets: []string{"secret1"},
		},
		"filter key value match": {
			Options:       []ListOption{WithListLabelSelector("foo=bar")},
			Setup:         createDummySecret("secret1", "default", map[string]string{"foo": "bar"}),
			ExpectSecrets: []string{"secret1"},
		},
		"filter key value mismatch": {
			Options:       []ListOption{WithListLabelSelector("foo=bazz")},
			Setup:         createDummySecret("secret1", "default", map[string]string{"foo": "bar"}),
			ExpectSecrets: []string{},
		},
		"filter disallowed label": {
			Options:       []ListOption{WithListLabelSelector("!foo")},
			Setup:         createDummySecret("secret1", "default", map[string]string{"foo": "bar"}),
			ExpectSecrets: []string{},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {

			mockK8s := testclient.NewSimpleClientset()
			k8sFactory := func() (kubernetes.Interface, error) {
				return mockK8s, tc.FactoryError
			}

			if tc.Setup != nil {
				tc.Setup(mockK8s)
			}

			secretsClient := NewClient(k8sFactory)
			secrets, actualErr := secretsClient.List(tc.Options...)
			if tc.ExpectErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectErr, actualErr)

				return
			}

			actualSecrets := []string{}
			for _, secret := range secrets {
				actualSecrets = append(actualSecrets, secret.Name)
			}

			sort.Strings(actualSecrets)
			sort.Strings(tc.ExpectSecrets)

			testutil.AssertEqual(t, "secrets", tc.ExpectSecrets, actualSecrets)
		})
	}
}
