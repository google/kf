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

package main

import (
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/buildpack/libbuildpack/detect"
	"github.com/cloudfoundry/archive-expanding-cnb/expand"
	"github.com/cloudfoundry/jvm-application-cnb/jvmapplication"
	"github.com/cloudfoundry/libcfbuildpack/test"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestDetect(t *testing.T) {
	spec.Run(t, "Detect", func(t *testing.T, _ spec.G, it spec.S) {

		g := NewGomegaWithT(t)

		var f *test.DetectFactory

		it.Before(func() {
			f = test.NewDetectFactory(t)
		})

		it("fails with no archive", func() {
			g.Expect(d(f.Detect)).To(Equal(detect.FailStatusCode))
		})

		it("fails with more than one archive", func() {
			test.TouchFile(t, f.Detect.Application.Root, "test-1.jar")
			test.TouchFile(t, f.Detect.Application.Root, "test-2.jar")

			g.Expect(d(f.Detect)).To(Equal(detect.FailStatusCode))
		})

		it("fails with non-root archive", func() {
			test.TouchFile(t, f.Detect.Application.Root, "sub-directory", "test-1.jar")

			g.Expect(d(f.Detect)).To(Equal(detect.FailStatusCode))
		})

		it("passes with .jar", func() {
			test.TouchFile(t, f.Detect.Application.Root, "test.jar")

			g.Expect(d(f.Detect)).To(Equal(detect.PassStatusCode))
			g.Expect(f.Output).To(Equal(buildplan.BuildPlan{
				expand.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{
						expand.Archive: filepath.Join(f.Detect.Application.Root, "test.jar"),
					},
				},
				jvmapplication.Dependency: buildplan.Dependency{},
			}))
		})

		it("passes with .war", func() {
			test.TouchFile(t, f.Detect.Application.Root, "test.war")

			g.Expect(d(f.Detect)).To(Equal(detect.PassStatusCode))
			g.Expect(f.Output).To(Equal(buildplan.BuildPlan{
				expand.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{
						expand.Archive: filepath.Join(f.Detect.Application.Root, "test.war"),
					},
				},
				jvmapplication.Dependency: buildplan.Dependency{},
			}))
		})

		it("passes with .tar", func() {
			test.TouchFile(t, f.Detect.Application.Root, "test.tar")

			g.Expect(d(f.Detect)).To(Equal(detect.PassStatusCode))
			g.Expect(f.Output).To(Equal(buildplan.BuildPlan{
				expand.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{
						expand.Archive: filepath.Join(f.Detect.Application.Root, "test.tar"),
					},
				},
				jvmapplication.Dependency: buildplan.Dependency{},
			}))
		})

		it("passes with .tar.gz", func() {
			test.TouchFile(t, f.Detect.Application.Root, "test.tar.gz")

			g.Expect(d(f.Detect)).To(Equal(detect.PassStatusCode))
			g.Expect(f.Output).To(Equal(buildplan.BuildPlan{
				expand.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{
						expand.Archive: filepath.Join(f.Detect.Application.Root, "test.tar.gz"),
					},
				},
				jvmapplication.Dependency: buildplan.Dependency{},
			}))
		})

		it("passes with .tgz", func() {
			test.TouchFile(t, f.Detect.Application.Root, "test.tgz")

			g.Expect(d(f.Detect)).To(Equal(detect.PassStatusCode))
			g.Expect(f.Output).To(Equal(buildplan.BuildPlan{
				expand.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{
						expand.Archive: filepath.Join(f.Detect.Application.Root, "test.tgz"),
					},
				},
				jvmapplication.Dependency: buildplan.Dependency{},
			}))
		})

		it("passes with .zip", func() {
			test.TouchFile(t, f.Detect.Application.Root, "test.zip")

			g.Expect(d(f.Detect)).To(Equal(detect.PassStatusCode))
			g.Expect(f.Output).To(Equal(buildplan.BuildPlan{
				expand.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{
						expand.Archive: filepath.Join(f.Detect.Application.Root, "test.zip"),
					},
				},
				jvmapplication.Dependency: buildplan.Dependency{},
			}))
		})
	}, spec.Report(report.Terminal{}))
}
