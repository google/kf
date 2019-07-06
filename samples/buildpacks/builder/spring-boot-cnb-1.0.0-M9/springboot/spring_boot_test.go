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

package springboot_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/jvm-application-cnb/jvmapplication"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/cloudfoundry/spring-boot-cnb/springboot"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestSpringBoot(t *testing.T) {
	spec.Run(t, "SpringBoot", func(t *testing.T, when spec.G, it spec.S) {

		g := NewGomegaWithT(t)

		var f *test.BuildFactory

		it.Before(func() {
			f = test.NewBuildFactory(t)
		})

		when("NewSpringBoot", func() {

			it("returns false when no jvm-application", func() {
				test.WriteFile(t, filepath.Join(f.Build.Application.Root, "META-INF", "MANIFEST.MF"),
					`
Spring-Boot-Classes: test-classes
Spring-Boot-Lib: test-lib
Start-Class: test-start-class
Spring-Boot-Version: test-version`)

				_, ok, err := springboot.NewSpringBoot(f.Build)
				g.Expect(ok).To(BeFalse())
				g.Expect(err).NotTo(HaveOccurred())
			})

			it("returns false when no Spring-Boot-Version", func() {
				f.AddBuildPlan(jvmapplication.Dependency, buildplan.Dependency{})
				test.WriteFile(t, filepath.Join(f.Build.Application.Root, "META-INF", "MANIFEST.MF"), "")

				_, ok, err := springboot.NewSpringBoot(f.Build)
				g.Expect(ok).To(BeFalse())
				g.Expect(err).NotTo(HaveOccurred())
			})

			it("returns true when Spring-Boot-Version exists", func() {
				f.AddBuildPlan(jvmapplication.Dependency, buildplan.Dependency{})
				test.WriteFile(t, filepath.Join(f.Build.Application.Root, "META-INF", "MANIFEST.MF"),
					`
Spring-Boot-Classes: test-classes
Spring-Boot-Lib: test-lib
Start-Class: test-start-class
Spring-Boot-Version: test-version`)

				_, ok, err := springboot.NewSpringBoot(f.Build)
				g.Expect(ok).To(BeTrue())
				g.Expect(err).NotTo(HaveOccurred())
			})
		})

		it("contributes command", func() {
			f.AddBuildPlan(jvmapplication.Dependency, buildplan.Dependency{})
			test.TouchFile(t, filepath.Join(f.Build.Application.Root, "test-lib", "test.jar"))
			test.WriteFile(t, filepath.Join(f.Build.Application.Root, "META-INF", "MANIFEST.MF"),
				`
Spring-Boot-Classes: test-classes
Spring-Boot-Lib: test-lib
Start-Class: test-start-class
Spring-Boot-Version: test-version`)

			e, ok, err := springboot.NewSpringBoot(f.Build)
			g.Expect(ok).To(BeTrue())
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(e.Contribute()).To(Succeed())

			layer := f.Build.Layers.Layer("spring-boot")
			g.Expect(layer).To(test.HaveLayerMetadata(true, true, true))
			g.Expect(layer).To(test.HaveAppendPathSharedEnvironment("CLASSPATH", strings.Join([]string{
				filepath.Join(f.Build.Application.Root, "test-classes"),
				filepath.Join(f.Build.Application.Root, "test-lib", "test.jar"),
			}, string(filepath.ListSeparator))))

			command := "java -cp $CLASSPATH $JAVA_OPTS test-start-class"
			g.Expect(f.Build.Layers).To(test.HaveApplicationMetadata(layers.Metadata{
				Processes: []layers.Process{
					{"spring-boot", command},
					{"task", command},
					{"web", command},
				},
			}))
		})
	}, spec.Report(report.Terminal{}))
}
