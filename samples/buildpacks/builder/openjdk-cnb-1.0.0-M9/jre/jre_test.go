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

package jre_test

import (
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/cloudfoundry/openjdk-cnb/jre"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestJRE(t *testing.T) {
	spec.Run(t, "JRE", func(t *testing.T, _ spec.G, it spec.S) {

		g := NewGomegaWithT(t)

		var f *test.BuildFactory

		it.Before(func() {
			f = test.NewBuildFactory(t)
		})

		it("returns true if build plan exists", func() {
			f.AddDependency(jre.Dependency, filepath.Join("testdata", "stub-openjdk-jre.tar.gz"))
			f.AddBuildPlan(jre.Dependency, buildplan.Dependency{})

			_, ok, err := jre.NewJRE(f.Build)
			g.Expect(ok).To(BeTrue())
			g.Expect(err).NotTo(HaveOccurred())
		})

		it("returns false if build plan does not exist", func() {
			f := test.NewBuildFactory(t)

			_, ok, err := jre.NewJRE(f.Build)
			g.Expect(ok).To(BeFalse())
			g.Expect(err).NotTo(HaveOccurred())
		})

		it("contributes JRE to build", func() {
			f.AddDependency(jre.Dependency, filepath.Join("testdata", "stub-openjdk-jre.tar.gz"))
			f.AddBuildPlan(jre.Dependency, buildplan.Dependency{
				Metadata: buildplan.Metadata{jre.BuildContribution: true},
			})

			j, _, err := jre.NewJRE(f.Build)
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(j.Contribute()).To(Succeed())

			layer := f.Build.Layers.Layer("openjdk-jre")
			g.Expect(layer).To(test.HaveLayerMetadata(true, true, false))
			g.Expect(filepath.Join(layer.Root, "fixture-marker")).To(BeARegularFile())
			g.Expect(layer).To(test.HaveOverrideSharedEnvironment("JAVA_HOME", layer.Root))
		})

		it("contributes JRE to launch", func() {
			f.AddDependency(jre.Dependency, filepath.Join("testdata", "stub-openjdk-jre.tar.gz"))
			f.AddBuildPlan(jre.Dependency, buildplan.Dependency{
				Metadata: buildplan.Metadata{jre.LaunchContribution: true},
			})

			j, _, err := jre.NewJRE(f.Build)
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(j.Contribute()).To(Succeed())

			layer := f.Build.Layers.Layer("openjdk-jre")
			g.Expect(layer).To(test.HaveLayerMetadata(false, false, true))
			g.Expect(filepath.Join(layer.Root, "fixture-marker")).To(BeARegularFile())
			g.Expect(layer).To(test.HaveOverrideSharedEnvironment("JAVA_HOME", layer.Root))
		})
	}, spec.Report(report.Terminal{}))
}
