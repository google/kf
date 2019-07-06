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

package jdk_test

import (
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/cloudfoundry/openjdk-cnb/jdk"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestJDK(t *testing.T) {
	spec.Run(t, "JDK", func(t *testing.T, _ spec.G, it spec.S) {

		g := NewGomegaWithT(t)

		var f *test.BuildFactory

		it.Before(func() {
			f = test.NewBuildFactory(t)
		})

		it("returns true if build plan exists", func() {
			f.AddDependency(jdk.Dependency, filepath.Join("testdata", "stub-openjdk-jdk.tar.gz"))
			f.AddBuildPlan(jdk.Dependency, buildplan.Dependency{})

			_, ok, err := jdk.NewJDK(f.Build)
			g.Expect(ok).To(BeTrue())
			g.Expect(err).NotTo(HaveOccurred())
		})

		it("returns false if build plan does not exist", func() {
			_, ok, err := jdk.NewJDK(f.Build)
			g.Expect(ok).To(BeFalse())
			g.Expect(err).NotTo(HaveOccurred())
		})

		it("contributes JDK", func() {
			f.AddDependency(jdk.Dependency, filepath.Join("testdata", "stub-openjdk-jdk.tar.gz"))
			f.AddBuildPlan(jdk.Dependency, buildplan.Dependency{})

			j, _, err := jdk.NewJDK(f.Build)
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(j.Contribute()).To(Succeed())

			layer := f.Build.Layers.Layer("openjdk-jdk")
			g.Expect(layer).To(test.HaveLayerMetadata(true, true, false))
			g.Expect(filepath.Join(layer.Root, "fixture-marker")).To(BeARegularFile())
			g.Expect(layer).To(test.HaveOverrideBuildEnvironment("JAVA_HOME", layer.Root))
			g.Expect(layer).To(test.HaveOverrideBuildEnvironment("JDK_HOME", layer.Root))
		})
	}, spec.Report(report.Terminal{}))
}
