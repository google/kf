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

package expand_test

import (
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/archive-expanding-cnb/expand"
	"github.com/cloudfoundry/libcfbuildpack/test"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func TestExpand(t *testing.T) {
	spec.Run(t, "Expand", func(t *testing.T, _ spec.G, it spec.S) {

		g := NewGomegaWithT(t)

		var f *test.BuildFactory

		it.Before(func() {
			f = test.NewBuildFactory(t)
		})

		it("returns true if build plan does exist", func() {
			f.AddBuildPlan(expand.Dependency, buildplan.Dependency{
				Metadata: buildplan.Metadata{
					expand.Archive: "test-archive",
				},
			})

			_, ok, err := expand.NewExpand(f.Build)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(ok).To(BeTrue())
		})

		it("returns false if build plan does not exist", func() {
			_, ok, err := expand.NewExpand(f.Build)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(ok).To(BeFalse())
		})

		it("expands .jar", func() {
			a := filepath.Join(f.Build.Application.Root, "stub-archive.jar")

			test.CopyFile(t, filepath.Join("testdata", "stub-archive.jar"), a)

			f.AddBuildPlan(expand.Dependency, buildplan.Dependency{
				Metadata: buildplan.Metadata{expand.Archive: a},
			})

			e, ok, err := expand.NewExpand(f.Build)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(ok).To(BeTrue())

			g.Expect(e.Contribute()).To(Succeed())

			g.Expect(filepath.Join(f.Build.Application.Root, "fixture-marker")).To(BeAnExistingFile())
			g.Expect(a).NotTo(BeAnExistingFile())
		})

		it("expands .war", func() {
			a := filepath.Join(f.Build.Application.Root, "stub-archive.war")

			test.CopyFile(t, filepath.Join("testdata", "stub-archive.war"), a)

			f.AddBuildPlan(expand.Dependency, buildplan.Dependency{
				Metadata: buildplan.Metadata{expand.Archive: a},
			})

			e, ok, err := expand.NewExpand(f.Build)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(ok).To(BeTrue())

			g.Expect(e.Contribute()).To(Succeed())

			g.Expect(filepath.Join(f.Build.Application.Root, "fixture-marker")).To(BeAnExistingFile())
			g.Expect(a).NotTo(BeAnExistingFile())
		})

		it("expands .tar", func() {
			a := filepath.Join(f.Build.Application.Root, "stub-archive.tar")

			test.CopyFile(t, filepath.Join("testdata", "stub-archive.tar"), a)

			f.AddBuildPlan(expand.Dependency, buildplan.Dependency{
				Metadata: buildplan.Metadata{expand.Archive: a},
			})

			e, ok, err := expand.NewExpand(f.Build)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(ok).To(BeTrue())

			g.Expect(e.Contribute()).To(Succeed())

			g.Expect(filepath.Join(f.Build.Application.Root, "fixture-marker")).To(BeAnExistingFile())
			g.Expect(a).NotTo(BeAnExistingFile())
		})

		it("expands .tar.gz", func() {
			a := filepath.Join(f.Build.Application.Root, "stub-archive.tar.gz")

			test.CopyFile(t, filepath.Join("testdata", "stub-archive.tar.gz"), a)

			f.AddBuildPlan(expand.Dependency, buildplan.Dependency{
				Metadata: buildplan.Metadata{expand.Archive: a},
			})

			e, ok, err := expand.NewExpand(f.Build)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(ok).To(BeTrue())

			g.Expect(e.Contribute()).To(Succeed())

			g.Expect(filepath.Join(f.Build.Application.Root, "fixture-marker")).To(BeAnExistingFile())
			g.Expect(a).NotTo(BeAnExistingFile())
		})

		it("expands .tgz", func() {
			a := filepath.Join(f.Build.Application.Root, "stub-archive.tgz")

			test.CopyFile(t, filepath.Join("testdata", "stub-archive.tgz"), a)

			f.AddBuildPlan(expand.Dependency, buildplan.Dependency{
				Metadata: buildplan.Metadata{expand.Archive: a},
			})

			e, ok, err := expand.NewExpand(f.Build)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(ok).To(BeTrue())

			g.Expect(e.Contribute()).To(Succeed())

			g.Expect(filepath.Join(f.Build.Application.Root, "fixture-marker")).To(BeAnExistingFile())
			g.Expect(a).NotTo(BeAnExistingFile())
		})

		it("expands .zip", func() {
			a := filepath.Join(f.Build.Application.Root, "stub-archive.zip")

			test.CopyFile(t, filepath.Join("testdata", "stub-archive.zip"), a)

			f.AddBuildPlan(expand.Dependency, buildplan.Dependency{
				Metadata: buildplan.Metadata{expand.Archive: a},
			})

			e, ok, err := expand.NewExpand(f.Build)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(ok).To(BeTrue())

			g.Expect(e.Contribute()).To(Succeed())

			g.Expect(filepath.Join(f.Build.Application.Root, "fixture-marker")).To(BeAnExistingFile())
			g.Expect(a).NotTo(BeAnExistingFile())
		})
	})
}
