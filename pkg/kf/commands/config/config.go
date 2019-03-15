package config

import "io"

// KfParams stores everything needed to interact with the user and Knative.
type KfParams struct {
	Output    io.Writer
	Namespace string
}
