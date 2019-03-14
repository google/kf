package secrets

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	testclient "k8s.io/client-go/kubernetes/fake"
)

func TestClient_Create(t *testing.T) {
	cases := map[string]struct {
		Name             string
		Options          []CreateOption
		CreateErr        error
		ExpectErr        error
		ExpectNamespace  string
		ExpectStringData map[string]string
		ExpectData       map[string][]byte
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
				if fmt.Sprint(tc.ExpectErr) != fmt.Sprint(actualErr) {
					t.Fatalf("wanted err: %v, got: %v", tc.ExpectErr, actualErr)
				}

				return
			}

			secret, err := mockK8s.CoreV1().Secrets(tc.ExpectNamespace).Get(tc.Name, v1.GetOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(secret.StringData, tc.ExpectStringData) {
				t.Errorf("Expected StringData %#v got: %#v", tc.ExpectStringData, secret.StringData)
			}

			if !reflect.DeepEqual(secret.Data, tc.ExpectData) {
				t.Errorf("Expected data %#v got: %#v", tc.ExpectData, secret.Data)
			}
		})
	}
}

func TestClient_Get(t *testing.T) {
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
				if fmt.Sprint(tc.ExpectErr) != fmt.Sprint(actualErr) {
					t.Fatalf("wanted err: %v, got: %v", tc.ExpectErr, actualErr)
				}

				return
			}

			if !reflect.DeepEqual(tc.ExpectSecrets, actualSecrets) {
				t.Errorf("wanted secrets: %v, got: %v", tc.ExpectSecrets, actualSecrets)
			}
		})
	}
}

func TestClient_Delete(t *testing.T) {
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
				if fmt.Sprint(tc.ExpectErr) != fmt.Sprint(actualErr) {
					t.Fatalf("wanted err: %v, got: %v", tc.ExpectErr, actualErr)
				}

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
