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
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/build-system-cnb/buildsystem"
	"github.com/cloudfoundry/build-system-cnb/runner"
	"github.com/cloudfoundry/libcfbuildpack/test"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestMaven(t *testing.T) {
	spec.Run(t, "Maven", func(t *testing.T, when spec.G, it spec.S) {

		g := NewGomegaWithT(t)

		var f *test.BuildFactory

		it.Before(func() {
			f = test.NewBuildFactory(t)

			f.AddDependency(buildsystem.MavenDependency, filepath.Join("testdata", "stub-maven.tar.gz"))
			f.AddBuildPlan(buildsystem.MavenDependency, buildplan.Dependency{})
			test.TouchFile(t, f.Build.Application.Root, ".mvn")
			test.TouchFile(t, f.Build.Application.Root, "mvnw")
		})

		when("working with JAR file", func() {
			it.Before(func() {
				test.CopyFile(t, filepath.Join("testdata", "stub-application.jar"),
					filepath.Join(f.Build.Application.Root, "target", "stub-application.jar"))
			})

			it("builds application", func() {
				f.Runner.Outputs = []string{"test-java-version"}

				b, _, err := buildsystem.NewMavenBuildSystem(f.Build)
				g.Expect(err).NotTo(HaveOccurred())
				r := runner.NewMavenRunner(f.Build, b)

				g.Expect(r.Contribute()).To(Succeed())

				g.Expect(f.Runner.Commands[1]).
					To(Equal(test.Command{
						Bin:  filepath.Join(f.Build.Application.Root, "mvnw"),
						Dir:  f.Build.Application.Root,
						Args: []string{"-Dmaven.test.skip=true", "package"},
					}))
			})

			it("removes source code", func() {
				f.Runner.Outputs = []string{"test-java-version"}

				b, _, err := buildsystem.NewMavenBuildSystem(f.Build)
				g.Expect(err).NotTo(HaveOccurred())
				r := runner.NewMavenRunner(f.Build, b)

				g.Expect(r.Contribute()).To(Succeed())

				g.Expect(f.Build.Application.Root).To(BeADirectory())
				g.Expect(f.Build.Application.Root).To(BeADirectory())
				g.Expect(filepath.Join(f.Build.Application.Root, ".mvn")).NotTo(BeAnExistingFile())
				g.Expect(filepath.Join(f.Build.Application.Root, "mvnw")).NotTo(BeAnExistingFile())
				g.Expect(filepath.Join(f.Build.Application.Root, "target")).NotTo(BeAnExistingFile())
			})

			it("explodes built application", func() {
				f.Runner.Outputs = []string{"test-java-version"}

				b, _, err := buildsystem.NewMavenBuildSystem(f.Build)
				g.Expect(err).NotTo(HaveOccurred())
				r := runner.NewMavenRunner(f.Build, b)

				g.Expect(r.Contribute()).To(Succeed())

				layer := f.Build.Layers.Layer("build-system-application")
				g.Expect(layer).To(test.HaveLayerMetadata(false, false, false))
				g.Expect(filepath.Join(f.Build.Application.Root, "fixture-marker")).To(BeARegularFile())
			})
		})

		when("working with WAR file", func() {
			it.Before(func() {
				test.CopyFile(t, filepath.Join("testdata", "stub-application.war"),
					filepath.Join(f.Build.Application.Root, "target", "stub-application.war"))
			})

			it("explodes built application", func() {
				f.Runner.Outputs = []string{"test-java-version"}

				b, _, err := buildsystem.NewMavenBuildSystem(f.Build)
				g.Expect(err).NotTo(HaveOccurred())
				r := runner.NewMavenRunner(f.Build, b)

				g.Expect(r.Contribute()).To(Succeed())

				layer := f.Build.Layers.Layer("build-system-application")
				g.Expect(layer).To(test.HaveLayerMetadata(false, false, false))
				g.Expect(filepath.Join(f.Build.Application.Root, "fixture-marker")).To(BeARegularFile())
			})
		})

	}, spec.Report(report.Terminal{}))
}
