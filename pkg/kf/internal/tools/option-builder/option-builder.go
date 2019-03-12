package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"strings"
	"text/template"
)

type Option struct {
	Name        string
	Type        string
	Description string
}

type OptionsConfig struct {
	Name    string
	Options []Option
	Imports []string

	// ConfigName is set by Name. It is modified to ensure its not exported.
	ConfigName string
}

func main() {
	common := []Option{
		{Name: "Namespace", Type: "string", Description: "the Kubertes namespace to use"},
	}

	configs := []OptionsConfig{
		{
			Name:    "Push",
			Imports: []string{"io"},
			Options: append(common, []Option{
				{Name: "Path", Type: "string", Description: "the path of the directory to push"},
				{Name: "ContainerRegistry", Type: "string", Description: "the container registry's URL"},
				{Name: "ServiceAccount", Type: "string", Description: "the service account to authenticate with"},
				{Name: "Output", Type: "io.Writer", Description: "the io.Writer to write output such as build logs"},
			}...),
		},
		{
			Name:    "Delete",
			Options: common,
		},
		{
			Name:    "List",
			Options: common,
		},
	}

	testTemplate := template.Must(template.New("").Funcs(template.FuncMap{}).Parse(`
// This file was generated with option-builder.go, DO NOT EDIT IT.

package kf
{{ if .Imports }}
import ({{ range $index, $import := .Imports }}{{ printf "\n\t%q" $import }}{{ end }}{{printf "\n"}})
{{ end }}

{{ $typecfg := (printf "%sConfig" .ConfigName) }}
type {{ $typecfg }} struct {
{{- range $i, $opt := .Options}}// {{ $opt.Name }} is {{ $opt.Description }}
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

`))

	for _, config := range configs {
		// Don't export the config name.
		config.ConfigName = strings.ToLower(string(config.Name[0])) + config.Name[1:]

		if err := generateCode(config, testTemplate); err != nil {
			panic(err)
		}
	}
}

func generateCode(config OptionsConfig, testTemplate *template.Template) error {
	f, err := os.Create(strings.ToLower(fmt.Sprintf("%s_options.go", config.Name)))
	if err != nil {
		return err
	}
	defer f.Close()

	buf := &bytes.Buffer{}
	if err := testTemplate.Execute(buf, config); err != nil {
		return err
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	f.Write(formatted)

	return nil
}
