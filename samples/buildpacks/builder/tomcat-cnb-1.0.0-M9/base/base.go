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

package base

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/buildpack/libbuildpack/application"
	"github.com/cloudfoundry/jvm-application-cnb/jvmapplication"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/buildpack"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
)

const (
	// AccessLoggingSupportDependency is the id for the Tomcat Access Logging Support contributed to the Tomcat instance.
	AccessLoggingSupportDependency = "tomcat-access-logging-support"

	// ExternalConfiguration is the id for the optional externalized configuration contributed to the Tomcat instance.
	ExternalConfiguration = "tomcat-external-configuration"

	// LifecycleSupportDependency is the id for the Tomcat Lifecycle Support contributed to the Tomcat instance.
	LifecycleSupportDependency = "tomcat-lifecycle-support"

	// LoggingSupportDependency is the id for the Tomcat Logging Support contributed to the Tomcat instance.
	LoggingSupportDependency = "tomcat-logging-support"
)

type Base struct {
	application application.Application
	buildpack   buildpack.Buildpack
	layer       layers.Layer

	contextPath                string
	dependencies               []buildpack.Dependency
	accessLoggingLayer         layers.DownloadLayer
	lifecycleLayer             layers.DownloadLayer
	loggingLayer               layers.DownloadLayer
	externalConfigurationLayer layers.DownloadLayer
}

func (b Base) Contribute() error {
	b.accessLoggingLayer.Touch()
	b.lifecycleLayer.Touch()
	b.loggingLayer.Touch()

	return b.layer.Contribute(marker{b.contextPath, b.dependencies}, func(layer layers.Layer) error {
		if err := os.RemoveAll(layer.Root); err != nil {
			return err
		}

		if err := b.contributeConfiguration(layer); err != nil {
			return err
		}

		if err := b.contributeAccessLogging(layer); err != nil {
			return err
		}

		if err := b.contributeLifecycleSupport(layer); err != nil {
			return err
		}

		if err := b.contributeLoggingSupport(layer); err != nil {
			return err
		}

		if err := b.contributeExternalConfiguration(layer); err != nil {
			return err
		}

		if err := b.contributeApplication(layer); err != nil {
			return err
		}

		return layer.OverrideLaunchEnv("CATALINA_BASE", layer.Root)
	}, layers.Launch)
}

func (b Base) contributeAccessLogging(layer layers.Layer) error {
	layer.Logger.Header("Contributing Access Logging Support")

	artifact, err := b.accessLoggingLayer.Artifact()
	if err != nil {
		return err
	}

	layer.Logger.Body("Copying to %s/lib", layer.Root)
	if err := helper.CopyFile(artifact, filepath.Join(layer.Root, "lib", filepath.Base(artifact))); err != nil {
		return err
	}

	return layer.WriteProfile("access-logging", `ENABLED=${BPL_TOMCAT_ACCESS_LOGGING:=n}

if [[ "${ENABLED}" = "n" ]]; then
	return
fi

printf "Tomcat Access Logging enabled\n"

export JAVA_OPTS="${JAVA_OPTS} -Daccess.logging.enabled=enabled"
`)
}

func (b Base) contributeApplication(layer layers.Layer) error {
	cp := filepath.Join(layer.Root, "webapps", b.contextPath)

	layer.Logger.Header("Mounting application at %s", cp)

	return helper.WriteSymlink(b.application.Root, cp)
}

func (b Base) contributeConfiguration(layer layers.Layer) error {
	layer.Logger.Header("Contributing Configuration")

	layer.Logger.Body("Copying context.xml to %s/conf", layer.Root)
	if err := helper.CopyFile(filepath.Join(b.buildpack.Root, "context.xml"), filepath.Join(layer.Root, "conf", "context.xml")); err != nil {
		return err
	}

	layer.Logger.Body("Copying web.xml to %s/conf", layer.Root)
	if err := helper.CopyFile(filepath.Join(b.buildpack.Root, "web.xml"), filepath.Join(layer.Root, "conf", "web.xml")); err != nil {
		return err
	}

	layer.Logger.Body("Copying logging.properties to %s/conf", layer.Root)
	if err := helper.CopyFile(filepath.Join(b.buildpack.Root, "logging.properties"), filepath.Join(layer.Root, "conf", "logging.properties")); err != nil {
		return err
	}

	layer.Logger.Body("Copying server.xml to %s/conf", layer.Root)
	if err := helper.CopyFile(filepath.Join(b.buildpack.Root, "server.xml"), filepath.Join(layer.Root, "conf", "server.xml")); err != nil {
		return err
	}

	return nil
}

func (b Base) contributeExternalConfiguration(layer layers.Layer) error {
	if reflect.DeepEqual(b.externalConfigurationLayer, layers.DownloadLayer{}) {
		return nil
	}

	layer.Logger.Header("Contributing External Configuration")

	artifact, err := b.externalConfigurationLayer.Artifact()
	if err != nil {
		return err
	}

	layer.Logger.Body("Expanding to %s", layer.Root)

	return helper.ExtractTarGz(artifact, layer.Root, 0)
}

func (b Base) contributeLifecycleSupport(layer layers.Layer) error {
	layer.Logger.Header("Contributing Lifecycle Support")

	artifact, err := b.lifecycleLayer.Artifact()
	if err != nil {
		return err
	}

	layer.Logger.Body("Copying to %s/lib", layer.Root)
	return helper.CopyFile(artifact, filepath.Join(layer.Root, "lib", filepath.Base(artifact)))
}

func (b Base) contributeLoggingSupport(layer layers.Layer) error {
	layer.Logger.Header("Contributing Logging Support")

	artifact, err := b.loggingLayer.Artifact()
	if err != nil {
		return err
	}

	destination := filepath.Join(layer.Root, "bin", filepath.Base(artifact))

	layer.Logger.Body("Copying to %s/bin", layer.Root)
	if err := helper.CopyFile(artifact, destination); err != nil {
		return err
	}

	layer.Logger.Body("Writing %s/bin/setenv.sh", layer.Root)
	return helper.WriteFile(filepath.Join(layer.Root, "bin", "setenv.sh"), 0755, `#!/bin/sh

CLASSPATH=%s`, destination)
}

type marker struct {
	ContextPath  string                 `toml:"context-path"`
	Dependencies []buildpack.Dependency `toml:"dependencies"`
}

func (m marker) Identity() (string, string) {
	return "Apache Tomcat Support", m.Dependencies[0].Version.Original()
}

// NewBase creates a new CATALINA_BASE instance.  OK is true if the application contains a "WEB-INF" directory.
func NewBase(build build.Build) (Base, bool, error) {
	if _, ok := build.BuildPlan[jvmapplication.Dependency]; !ok {
		return Base{}, false, nil
	}

	ok, err := helper.FileExists(filepath.Join(build.Application.Root, "WEB-INF"))
	if err != nil {
		return Base{}, false, err
	}

	if !ok {
		return Base{}, false, nil
	}

	deps, err := build.Buildpack.Dependencies()
	if err != nil {
		return Base{}, false, err
	}

	var d []buildpack.Dependency

	al, err := deps.Best(AccessLoggingSupportDependency, "", build.Stack)
	if err != nil {
		return Base{}, false, err
	}
	d = append(d, al)

	lc, err := deps.Best(LifecycleSupportDependency, "", build.Stack)
	if err != nil {
		return Base{}, false, err
	}
	d = append(d, lc)

	log, err := deps.Best(LoggingSupportDependency, "", build.Stack)
	if err != nil {
		return Base{}, false, err
	}
	d = append(d, log)

	var externalConfigurationLayer layers.DownloadLayer
	if deps.Has(ExternalConfiguration) {
		e, err := deps.Best(ExternalConfiguration, "", build.Stack)
		if err != nil {
			return Base{}, false, err
		}
		d = append(d, e)

		externalConfigurationLayer = build.Layers.DownloadLayer(e)
	}

	return Base{
		build.Application,
		build.Buildpack,
		build.Layers.Layer("catalina-base"),
		contextPath(),
		d,
		build.Layers.DownloadLayer(al),
		build.Layers.DownloadLayer(lc),
		build.Layers.DownloadLayer(log),
		externalConfigurationLayer,
	}, true, nil
}

func contextPath() string {
	cp, ok := os.LookupEnv("BP_TOMCAT_CONTEXT_PATH")
	if !ok {
		cp = "ROOT"
	}

	cp = regexp.MustCompile("^/").ReplaceAllString(cp, "")
	return strings.ReplaceAll(cp, "/", "#")
}
