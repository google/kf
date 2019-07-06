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

package base_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/jvm-application-cnb/jvmapplication"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/cloudfoundry/tomcat-cnb/base"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestBase(t *testing.T) {
	spec.Run(t, "Base", func(t *testing.T, when spec.G, it spec.S) {

		g := NewGomegaWithT(t)

		var f *test.BuildFactory

		it.Before(func() {
			f = test.NewBuildFactory(t)
		})

		it("results false with no jvm-application", func() {
			if err := os.MkdirAll(filepath.Join(f.Build.Application.Root, "WEB-INF"), 0755); err != nil {
				t.Fatal(err)
			}

			_, ok, err := base.NewBase(f.Build)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(ok).To(BeFalse())
		})

		it("returns false with no WEB-INF", func() {
			f.AddBuildPlan(jvmapplication.Dependency, buildplan.Dependency{})

			_, ok, err := base.NewBase(f.Build)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(ok).To(BeFalse())
		})

		when("valid application", func() {

			it.Before(func() {
				f.AddDependency("tomcat-access-logging-support", filepath.Join("testdata", "stub-tomcat-access-logging-support.jar"))
				f.AddDependency("tomcat-lifecycle-support", filepath.Join("testdata", "stub-tomcat-lifecycle-support.jar"))
				f.AddDependency("tomcat-logging-support", filepath.Join("testdata", "stub-tomcat-logging-support.jar"))
				f.AddBuildPlan(jvmapplication.Dependency, buildplan.Dependency{})
				test.TouchFile(t, filepath.Join(f.Build.Buildpack.Root, "context.xml"))
				test.TouchFile(t, filepath.Join(f.Build.Buildpack.Root, "logging.properties"))
				test.TouchFile(t, filepath.Join(f.Build.Buildpack.Root, "server.xml"))

				if err := os.MkdirAll(filepath.Join(f.Build.Application.Root, "WEB-INF"), 0755); err != nil {
					t.Fatal(err)
				}
			})

			it("returns true with jvm-application and WEB-INF", func() {
				_, ok, err := base.NewBase(f.Build)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(ok).To(BeTrue())
			})

			it("links application to ROOT", func() {
				b, _, err := base.NewBase(f.Build)
				g.Expect(err).NotTo(HaveOccurred())

				g.Expect(b.Contribute()).To(Succeed())

				layer := f.Build.Layers.Layer("catalina-base")
				g.Expect(filepath.Join(layer.Root, "webapps", "ROOT")).To(test.BeASymlink(f.Build.Application.Root))
			})

			it("links application to BP_TOMCAT_CONTEXT_PATH", func() {
				defer test.ReplaceEnv(t, "BP_TOMCAT_CONTEXT_PATH", "foo/bar")()

				b, _, err := base.NewBase(f.Build)
				g.Expect(err).NotTo(HaveOccurred())

				g.Expect(b.Contribute()).To(Succeed())

				layer := f.Build.Layers.Layer("catalina-base")
				g.Expect(filepath.Join(layer.Root, "webapps", "foo#bar")).To(test.BeASymlink(f.Build.Application.Root))
			})

			it("contributes configuration", func() {
				b, _, err := base.NewBase(f.Build)
				g.Expect(err).NotTo(HaveOccurred())

				g.Expect(b.Contribute()).To(Succeed())

				layer := f.Build.Layers.Layer("catalina-base")
				g.Expect(filepath.Join(layer.Root, "conf", "context.xml")).To(BeAnExistingFile())
				g.Expect(filepath.Join(layer.Root, "conf", "logging.properties")).To(BeAnExistingFile())
				g.Expect(filepath.Join(layer.Root, "conf", "server.xml")).To(BeAnExistingFile())
			})

			it("contributes access logging support", func() {
				b, _, err := base.NewBase(f.Build)
				g.Expect(err).NotTo(HaveOccurred())

				g.Expect(b.Contribute()).To(Succeed())

				layer := f.Build.Layers.Layer("catalina-base")
				g.Expect(filepath.Join(layer.Root, "lib", "stub-tomcat-access-logging-support.jar")).To(BeAnExistingFile())
				g.Expect(layer).To(test.HaveProfile("access-logging", `ENABLED=${BPL_TOMCAT_ACCESS_LOGGING:=n}

if [[ "${ENABLED}" = "n" ]]; then
	return
fi

printf "Tomcat Access Logging enabled\n"

export JAVA_OPTS="${JAVA_OPTS} -Daccess.logging.enabled=enabled"
`))
			})

			it("contributes lifecycle support", func() {
				b, _, err := base.NewBase(f.Build)
				g.Expect(err).NotTo(HaveOccurred())

				g.Expect(b.Contribute()).To(Succeed())

				layer := f.Build.Layers.Layer("catalina-base")
				g.Expect(filepath.Join(layer.Root, "lib", "stub-tomcat-lifecycle-support.jar")).To(BeAnExistingFile())
			})

			it("contributes logging support", func() {
				b, _, err := base.NewBase(f.Build)
				g.Expect(err).NotTo(HaveOccurred())

				g.Expect(b.Contribute()).To(Succeed())

				layer := f.Build.Layers.Layer("catalina-base")
				destination := filepath.Join(layer.Root, "bin", "stub-tomcat-logging-support.jar")
				g.Expect(destination).To(BeAnExistingFile())
				g.Expect(filepath.Join(layer.Root, "bin", "setenv.sh")).To(test.HavePermissions(0755))
				g.Expect(filepath.Join(layer.Root, "bin", "setenv.sh")).To(test.HaveContent(fmt.Sprintf(`#!/bin/sh

CLASSPATH=%s`, destination)))
			})

			it("contributes external configuration", func() {
				f.AddDependency("tomcat-external-configuration", filepath.Join("testdata", "stub-external-configuration.tar.gz"))

				b, _, err := base.NewBase(f.Build)
				g.Expect(err).NotTo(HaveOccurred())

				g.Expect(b.Contribute()).To(Succeed())

				layer := f.Build.Layers.Layer("catalina-base")
				g.Expect(filepath.Join(layer.Root, "fixture-marker")).To(BeAnExistingFile())
			})

			it("sets CATALINA_BASE", func() {
				b, _, err := base.NewBase(f.Build)
				g.Expect(err).NotTo(HaveOccurred())

				g.Expect(b.Contribute()).To(Succeed())

				layer := f.Build.Layers.Layer("catalina-base")
				g.Expect(layer).To(test.HaveLayerMetadata(false, false, true))
				g.Expect(layer).To(test.HaveOverrideLaunchEnvironment("CATALINA_BASE", layer.Root))
			})
		})
	}, spec.Report(report.Terminal{}))
}
