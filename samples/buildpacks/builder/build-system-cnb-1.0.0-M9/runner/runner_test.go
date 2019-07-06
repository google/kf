/*
 * Copyright 2018-2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package runner_test

import (
	"os/exec"

	"github.com/buildpack/libbuildpack/application"
	"github.com/cloudfoundry/libcfbuildpack/logger"
)

type testExecutor struct {
	Commands []*exec.Cmd
	Outputs  []string
}

func (t *testExecutor) Execute(application application.Application, cmd *exec.Cmd, logger logger.Logger) error {
	t.Commands = append(t.Commands, cmd)
	return nil
}

func (t *testExecutor) ExecuteWithOutput(application application.Application, cmd *exec.Cmd, logger logger.Logger) ([]byte, error) {
	t.Commands = append(t.Commands, cmd)

	output := t.Outputs[0]
	t.Outputs = t.Outputs[1:]

	return []byte(output), nil
}
