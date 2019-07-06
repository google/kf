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

package cache_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/build-system-cnb/cache"
	"github.com/cloudfoundry/libcfbuildpack/test"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestCache(t *testing.T) {
	spec.Run(t, "Cache", func(t *testing.T, _ spec.G, it spec.S) {

		g := NewGomegaWithT(t)

		it("contributes destination if it does not exist", func() {
			f := test.NewBuildFactory(t)

			destination := filepath.Join(f.Home, "target")

			c, err := cache.NewCache(f.Build, destination)
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(c.Contribute()).To(Succeed())

			layer := f.Build.Layers.Layer("build-system-cache")
			g.Expect(layer).To(test.HaveLayerMetadata(false, true, false))
			g.Expect(destination).To(test.BeASymlink(layer.Root))
		})
	}, spec.Report(report.Terminal{}))
}
