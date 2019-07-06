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

func TestGradle(t *testing.T) {
	spec.Run(t, "Gradle", func(t *testing.T, when spec.G, it spec.S) {

		g := NewGomegaWithT(t)

		var f *test.BuildFactory

		it.Before(func() {
			f = test.NewBuildFactory(t)
		})

		it("contains gradle, jvm-application, and openjdk-jdk in build plan", func() {
			g.Expect(buildsystem.GradleBuildPlanContribution(f.Build.BuildPlan)).To(Equal(buildplan.BuildPlan{
				buildsystem.GradleDependency: buildplan.Dependency{},
				jvmapplication.Dependency:    buildplan.Dependency{},
				jdk.Dependency:               buildplan.Dependency{},
			}))
		})

		when("Contribute", func() {

			it("contributes gradle if gradlew does not exist", func() {
				f.AddDependency(buildsystem.GradleDependency, filepath.Join("testdata", "stub-gradle.zip"))
				f.AddBuildPlan(buildsystem.GradleDependency, buildplan.Dependency{})

				b, _, err := buildsystem.NewGradleBuildSystem(f.Build)
				g.Expect(err).NotTo(HaveOccurred())

				g.Expect(b.Contribute()).To(Succeed())

				layer := f.Build.Layers.Layer("gradle")
				g.Expect(layer).To(test.HaveLayerMetadata(false, true, false))
				g.Expect(filepath.Join(layer.Root, "fixture-marker")).To(BeARegularFile())
			})

			it("does not contribute gradle if gradlew does exist", func() {
				f.AddDependency(buildsystem.GradleDependency, filepath.Join("testdata", "stub-gradle.zip"))
				f.AddBuildPlan(buildsystem.GradleDependency, buildplan.Dependency{})

				test.TouchFile(t, f.Build.Application.Root, "gradlew")

				b, _, err := buildsystem.NewGradleBuildSystem(f.Build)
				g.Expect(err).NotTo(HaveOccurred())

				g.Expect(b.Contribute()).To(Succeed())

				layer := f.Build.Layers.Layer("gradle")
				g.Expect(filepath.Join(layer.Root, "fixture-marker")).NotTo(BeAnExistingFile())
			})
		})

		when("IsGradle", func() {

			it("returns false if build.gradle does not exist", func() {
				g.Expect(buildsystem.IsGradle(f.Build.Application)).To(BeFalse())
			})

			it("returns true if build.gradle does exist", func() {
				test.TouchFile(t, f.Build.Application.Root, "build.gradle")

				g.Expect(buildsystem.IsGradle(f.Build.Application)).To(BeTrue())
			})

			it("returns true if build.gradle.kts does exist", func() {
				test.TouchFile(t, f.Build.Application.Root, "build.gradle.kts")

				g.Expect(buildsystem.IsGradle(f.Build.Application)).To(BeTrue())
			})
		})

		when("NewGradleBuildSystem", func() {

			it("returns true if build plan exists", func() {
				f.AddDependency(buildsystem.GradleDependency, filepath.Join("testdata", "stub-gradle.zip"))
				f.AddBuildPlan(buildsystem.GradleDependency, buildplan.Dependency{})

				_, ok, err := buildsystem.NewGradleBuildSystem(f.Build)
				g.Expect(ok).To(BeTrue())
				g.Expect(err).NotTo(HaveOccurred())
			})

			it("returns false if build plan does not exist", func() {
				_, ok, err := buildsystem.NewGradleBuildSystem(f.Build)
				g.Expect(ok).To(BeFalse())
				g.Expect(err).NotTo(HaveOccurred())
			})
		})
	}, spec.Report(report.Terminal{}))
}
