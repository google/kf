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

	"github.com/cloudfoundry/libcfbuildpack/test"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestGoogleStackdriverCredentials(t *testing.T) {
	spec.Run(t, "Google Stackdriver Credentials", func(t *testing.T, when spec.G, it spec.S) {

		g := NewGomegaWithT(t)

		var root string

		it.Before(func() {
			root = test.ScratchDir(t, "google-stackdriver-credentials")
		})

		when("debugger service is bound", func() {

			it("returns instrumentation key", func() {
				defer test.ReplaceEnv(t, "CNB_SERVICES", `{
  "google-stackdriver-debugger": [
    {
      "credentials": {
		"PrivateKeyData": "eyJwcm9qZWN0X2lkIjoidGVzdC1wcm9qZWN0LWlkIn0="
      },
      "label": "google-stackdriver-debugger"
    }
  ]
}
`)()

				f := filepath.Join(root, "credentials.json")

				code, err := p(f)

				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(code).To(Equal(0))
				g.Expect(f).To(test.HaveContent(`{"project_id":"test-project-id"}`))
			})
		})

		when("profiler service is bound", func() {

			it("returns instrumentation key", func() {
				defer test.ReplaceEnv(t, "CNB_SERVICES", `{
  "google-stackdriver-profiler": [
    {
      "credentials": {
		"PrivateKeyData": "eyJwcm9qZWN0X2lkIjoidGVzdC1wcm9qZWN0LWlkIn0="
      },
      "label": "google-stackdriver-profiler"
    }
  ]
}
`)()

				f := filepath.Join(root, "credentials.json")

				code, err := p(f)

				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(code).To(Equal(0))
				g.Expect(f).To(test.HaveContent(`{"project_id":"test-project-id"}`))
			})
		})

		when("service is not bound", func() {

			it("returns empty", func() {
				f := filepath.Join(root, "credentials.json")

				code, err := p(f)

				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(code).To(Equal(0))
				g.Expect(f).NotTo(BeAnExistingFile())
			})
		})
	}, spec.Report(report.Terminal{}))
}
