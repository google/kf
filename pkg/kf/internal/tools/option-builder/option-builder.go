// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build ignore

package main

import (
	"bytes"
	"flag"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/google/kf/pkg/kf/internal/tools/generator"
	"gopkg.in/yaml.v2"
)

type Option struct {
	Name        string  `yaml:"name"`
	Type        string  `yaml:"type"`
	Description string  `yaml:"description"`
	Default     *string `yaml:"default"`
}

type OptionsConfig struct {
	Name    string   `yaml:"name"`
	Options []Option `yaml:"options"`

	// ConfigName is set by Name. It is modified to ensure its not exported.
	ConfigName string `yaml:"-"`
}

type OptionsFile struct {
	License string            `yaml:"license"`
	Package string            `yaml:"package"`
	Imports map[string]string `yaml:"imports"`

	CommonOptions []Option        `yaml:"common"`
	Configs       []OptionsConfig `yaml:"configs"`
}

func main() {
	log.SetPrefix("option-builder ")
	log.Println("starting")

	pkg := flag.String("pkg", "", "overrides the package")
	flag.Parse()
	args := flag.Args()

	if flag.NArg() != 2 {
		log.Fatalln("use: option-builder.go /path/to/options.yml output-file.go")
	}

	optionsPath := args[0]
	optionsOut := args[1]

	log.Printf("building %s from %s\n", optionsOut, optionsPath)

	contents, err := ioutil.ReadFile(optionsPath)
	if err != nil {
		log.Fatalln(err)
	}

	of := &OptionsFile{}
	if err := yaml.Unmarshal(contents, of); err != nil {
		log.Fatalln(err)
	}

	if *pkg != "" {
		of.Package = *pkg
	}

	var configs []OptionsConfig
	for _, cfg := range of.Configs {
		// Mix in the common options and sort by name.
		cfg.Options = append(of.CommonOptions, cfg.Options...)
		sort.Slice(cfg.Options, func(i, j int) bool {
			return cfg.Options[i].Name < cfg.Options[j].Name
		})

		// Don't export the config name.
		cfg.ConfigName = strings.ToLower(string(cfg.Name[0])) + cfg.Name[1:]

		configs = append(configs, cfg)
	}

	// TODO: Read in LICENSE-HEADER file
	headerTemplate := template.Must(template.New("").Funcs(generator.TemplateFuncs()).Parse(`
{{genlicense}}

{{gennotice "option-builder.go"}}

package {{.Package}}

{{genimports .Imports}}
`))

	testTemplate := template.Must(template.New("").Funcs(generator.TemplateFuncs()).Parse(`
{{ $typecfg := (printf "%sConfig" .ConfigName) }}
type {{ $typecfg }} struct {
{{ range $i, $opt := .Options}}// {{ $opt.Name }} is {{ $opt.Description }}
{{ $opt.Name }}  {{ $opt.Type }}
{{ end }}
}

{{ $typeopt := (printf "%sOption" .Name) }}
// {{ $typeopt }} is a single option for configuring a {{.ConfigName}}Config
type {{.Name}}Option func(*{{.ConfigName}}Config)

{{ $typeoptarr := (printf "%sOptions" .Name) }}
// {{ $typeoptarr }} is a configuration set defining a {{.ConfigName}}Config
type {{ $typeoptarr }} []{{ $typeopt }}

// toConfig applies all the options to a new {{ $typecfg }} and returns it.
func (opts {{ $typeoptarr }}) toConfig() {{ $typecfg }} {
  cfg := {{ $typecfg }}{}

  for _, v := range opts {
    v(&cfg)
  }

  return cfg
}

// Extend creates a new {{ $typeoptarr }} with the contents of other overriding
// the values set in this {{ $typeoptarr }}.
func (opts {{ $typeoptarr }}) Extend(other {{ $typeoptarr }}) {{ $typeoptarr }} {
	var out {{ $typeoptarr }}
	out = append(out, opts...)
	out = append(out, other...)
	return out
}

{{ $globalName := .Name }}
{{ range $i, $opt := .Options}}

// {{$opt.Name}} returns the last set value for {{$opt.Name}} or the empty value
// if not set.
func (opts {{ $typeoptarr }}) {{$opt.Name}}() {{ $opt.Type }} {
  return opts.toConfig().{{$opt.Name}}
}

{{ end }}
{{ range $i, $opt := .Options}}

// With{{$globalName}}{{$opt.Name}} creates an Option that sets {{$opt.Description}}
func With{{$globalName}}{{$opt.Name}}(val {{ $opt.Type }}) {{ $typeopt }} {
  return func(cfg *{{ $typecfg }}) {
    cfg.{{$opt.Name}} = val
  }
}

{{ end }}

// {{$globalName}}OptionDefaults gets the default values for {{$globalName}}.
func {{$globalName}}OptionDefaults() {{ $typeoptarr }} {
	return {{ $typeoptarr }}{
		{{ range $i, $opt := .Options}}{{ if $opt.Default }}With{{$globalName}}{{$opt.Name}}({{ $opt.Default }}),
		{{ end }}{{ end }}
	}
}

`))

	f, err := os.Create(optionsOut)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	buf := &bytes.Buffer{}
	if err := headerTemplate.Execute(buf, of); err != nil {
		log.Fatalln(err)
	}

	buf.WriteString("\n")

	for _, config := range configs {
		if err := testTemplate.Execute(buf, config); err != nil {
			log.Fatalln(err)
		}
		buf.WriteString("\n")
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatalln(err)
	}

	f.Write(formatted)
}
