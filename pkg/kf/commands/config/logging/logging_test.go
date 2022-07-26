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

package logging_test

import (
	"context"
	"os"

	configlogging "github.com/google/kf/v2/pkg/kf/commands/config/logging"
	"go.uber.org/zap"
	"knative.dev/pkg/logging"
)

func ExampleSetupLogger() {
	logger := logging.FromContext(configlogging.SetupLogger(context.Background(), os.Stdout))

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warning message")
	logger.Error("error message")
	logger.With(zap.Namespace("some namespace")).Info("info message with namespace")
	logger.With(zap.String("some key", "some value")).Debug("debug message with key/value")
	logger.With(zap.Namespace("ns1")).With(zap.Namespace("ns2")).Info("info message with multiple namespaces")

	// Output: DEBUG debug message
	// info message
	// WARN warning message
	// ERROR error message
	// [some namespace] info message with namespace
	// DEBUG [some key/some value] debug message with key/value
	// [ns1] [ns2] info message with multiple namespaces
}
