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
	"testing"

	"github.com/buildpack/libbuildpack/detect"
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

		it("fails without build system", func() {
			g.Expect(d(f.Detect)).To(Equal(detect.FailStatusCode))
		})

		it("passes with build.gradle", func() {
			test.TouchFile(t, f.Detect.Application.Root, "build.gradle")

			g.Expect(d(f.Detect)).To(Equal(detect.PassStatusCode))
		})

		it("passes with build.gradle.kts", func() {
			test.TouchFile(t, f.Detect.Application.Root, "build.gradle.kts")

			g.Expect(d(f.Detect)).To(Equal(detect.PassStatusCode))
		})

		it("passes with pom.xml", func() {
			test.TouchFile(t, f.Detect.Application.Root, "pom.xml")

			g.Expect(d(f.Detect)).To(Equal(detect.PassStatusCode))
		})
	}, spec.Report(report.Terminal{}))
}
