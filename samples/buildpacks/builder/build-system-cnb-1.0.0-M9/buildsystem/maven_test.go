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

package buildsystem_test

import (
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/build-system-cnb/buildsystem"
	"github.com/cloudfoundry/jvm-application-cnb/jvmapplication"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/cloudfoundry/openjdk-cnb/jdk"
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
		})

		it("contains maven, jvm-application, and openjdk-jdk in build plan", func() {
			g.Expect(buildsystem.MavenBuildPlanContribution(f.Build.BuildPlan)).To(Equal(buildplan.BuildPlan{
				buildsystem.MavenDependency: buildplan.Dependency{},
				jvmapplication.Dependency:   buildplan.Dependency{},
				jdk.Dependency:              buildplan.Dependency{},
			}))
		})

		when("Contribute", func() {

			it("contributes maven if mvnw does not exist", func() {
				f.AddDependency(buildsystem.MavenDependency, filepath.Join("testdata", "stub-maven.tar.gz"))
				f.AddBuildPlan(buildsystem.MavenDependency, buildplan.Dependency{})

				b, _, err := buildsystem.NewMavenBuildSystem(f.Build)
				g.Expect(err).NotTo(HaveOccurred())

				g.Expect(b.Contribute()).To(Succeed())

				layer := f.Build.Layers.Layer("maven")
				g.Expect(layer).To(test.HaveLayerMetadata(false, true, false))
				g.Expect(filepath.Join(layer.Root, "fixture-marker")).To(BeARegularFile())
			})

			it("does not contribute maven if mvnw does exist", func() {
				f.AddDependency(buildsystem.MavenDependency, filepath.Join("testdata", "stub-maven.tar.gz"))
				f.AddBuildPlan(buildsystem.MavenDependency, buildplan.Dependency{})

				test.TouchFile(t, f.Build.Application.Root, "mvnw")

				b, _, err := buildsystem.NewMavenBuildSystem(f.Build)
				g.Expect(err).NotTo(HaveOccurred())

				g.Expect(b.Contribute()).To(Succeed())

				layer := f.Build.Layers.Layer("maven")
				g.Expect(filepath.Join(layer.Root, "fixture-marker")).NotTo(BeAnExistingFile())
			})
		})

		when("IsMaven", func() {

			it("returns false if pom.xml does not exist", func() {
				g.Expect(buildsystem.IsMaven(f.Build.Application)).To(BeFalse())
			})

			it("returns true if pom.xml does exist", func() {
				test.TouchFile(t, f.Build.Application.Root, "pom.xml")

				g.Expect(buildsystem.IsMaven(f.Build.Application)).To(BeTrue())
			})
		})

		when("NewMavenBuildSystem", func() {

			it("returns true if build plan exists", func() {
				f.AddDependency(buildsystem.MavenDependency, filepath.Join("testdata", "stub-maven.tar.gz"))
				f.AddBuildPlan(buildsystem.MavenDependency, buildplan.Dependency{})

				_, ok, err := buildsystem.NewMavenBuildSystem(f.Build)
				g.Expect(ok).To(BeTrue())
				g.Expect(err).NotTo(HaveOccurred())
			})

			it("returns false if build plan does not exist", func() {
				_, ok, err := buildsystem.NewMavenBuildSystem(f.Build)
				g.Expect(ok).To(BeFalse())
				g.Expect(err).NotTo(HaveOccurred())
			})
		})
	}, spec.Report(report.Terminal{}))
}
