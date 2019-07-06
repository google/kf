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

package cli_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/cloudfoundry/spring-boot-cnb/cli"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestCLI(t *testing.T) {
	spec.Run(t, "Spring Boot CLI", func(t *testing.T, when spec.G, it spec.S) {

		g := NewGomegaWithT(t)

		var f *test.BuildFactory

		it.Before(func() {
			f = test.NewBuildFactory(t)
		})

		it("contributes cli", func() {
			f.AddDependency(cli.Dependency, filepath.Join("testdata", "stub-spring-boot-cli.tar.gz"))

			a, err := cli.NewCLI(f.Build)
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(a.Contribute()).To(Succeed())

			layer := f.Build.Layers.Layer("spring-boot-cli")
			g.Expect(layer).To(test.HaveLayerMetadata(false, false, true))
			g.Expect(filepath.Join(layer.Root, "bin", "spring")).To(BeARegularFile())
		})
	}, spec.Report(report.Terminal{}))
}
