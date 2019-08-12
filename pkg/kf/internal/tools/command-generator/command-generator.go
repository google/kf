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

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/google/kf/pkg/kf/internal/tools/generator"
	"gopkg.in/yaml.v2"
)

type CommandFile struct {
	License string            `yaml:"license"`
	Package string            `yaml:"package"`
	Imports map[string]string `yaml:"imports"`

	Commands []Command `yaml:"commands"`
}

type Command struct {
	Name     string   `yaml:"name"`
	Short    string   `yaml:"short"`
	Long     string   `yaml:"long"`
	Aliases  []string `yaml:"aliases"`
	Examples []string `yaml:"examples"`

	IncludeManifest bool   `yaml:"includeManifest"`
	MarkdownPath    string `yaml:"markdownPath"`

	Args  []Arg  `yaml:"args"`
	Flags []Flag `yaml:"flags"`
}

func (c Command) Unexported() string {
	return generator.UnexportedName(camelCase(c.Name))
}

func (c Command) Exported() string {
	return generator.ExportedName(camelCase(c.Name))
}

func (c Command) Use() string {
	args := []string{}
	for _, arg := range c.Args {
		name := strings.ToUpper(arg.Name)
		if arg.Optional {
			name = fmt.Sprintf("[%s]", name)
		}
		args = append(args, name)
	}

	return c.Name + " " + strings.Join(args, " ")
}

// IndentedExample puts 2 spaces before each line of an example. This is
// necessary because every other element (e.g., usage, flags), are indented,
// but the example (if its multilined) is not.
func (c Command) IndentedExample() string {
	for i := range c.Examples {
		c.Examples[i] = "  " + c.Examples[i]
	}
	return strings.Join(c.Examples, "\n")
}

func (c Command) AliasStringArray() string {
	var x []string
	for _, a := range c.Aliases {
		x = append(x, fmt.Sprintf("%q", a))
	}
	return fmt.Sprintf("[]string { %s }", strings.Join(x, ","))
}

func (c Command) ArgNRange() string {
	var (
		required      int
		foundOptional bool
	)
	for _, arg := range c.Args {
		if arg.Optional {
			foundOptional = true
			continue
		}
		// Assert that optional args are after required.
		if foundOptional {
			log.Fatalf("%s is a REQUIRED arg, however it is after an OPTIONAL arg. This is invalid.", arg.Name)
		}

		required++
	}
	return fmt.Sprintf("%d, %d", required, len(c.Args))
}

type Arg struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Optional    bool   `yaml:"optional"`

	// UpperName is set by Name.
	UpperName string `yaml:"-"`
}

type Flag struct {
	Long      string `yaml:"long"`
	Type      string `yaml:"type"`
	Shorthand string `yaml:"shorthand"`
	Default   string `yaml:"default"`
	Usage     string `yaml:"usage"`
}

func (f Flag) Exported() string {
	return generator.ExportedName(camelCase(f.Long))
}

func (f Flag) Unexported() string {
	return generator.UnexportedName(camelCase(f.Long))
}

func (f Flag) CobraFuncName() string {
	return generator.ExportedName(f.Type) + "VarP"
}

func (f Flag) ManifestStructTag() string {
	return fmt.Sprintf("`yaml:%q`", generator.UnexportedName(camelCase(f.Long)))
}

func (f Flag) DefaultValue() (value string) {
	defer func() {
		// We need strings to be wrapped in quotes
		if f.Type != "string" {
			return
		}

		value = fmt.Sprintf("%q", value)
	}()

	if f.Default != "" {
		return f.Default
	}

	switch f.Type {
	case "int64", "int", "float64":
		return "0"
	case "string":
		return ""
	case "bool":
		return "false"
	default:
		panic("unkown type " + f.Type)
	}
}

func (f Flag) MarkdownTitle() string {
	switch {
	case f.Long != "" && f.Shorthand != "":
		return fmt.Sprintf("##### -%s, --%s <_%s_>", f.Shorthand, f.Long, f.Type)
	case f.Long != "":
		return fmt.Sprintf("##### --%s <_%s_>", f.Long, f.Type)
	case f.Shorthand != "":
		return fmt.Sprintf("##### -%s <_%s_>", f.Shorthand, f.Type)
	default:
		panic("Long and Short not set")
	}
}

func (f Flag) MarkdownDefault() string {
	if f.Default != "" {
		if f.Type == "string" {
			return fmt.Sprintf("%s (default=%q)", f.Usage, f.Default)
		}
		return fmt.Sprintf("%s (default=%s)", f.Usage, f.Default)
	}

	return fmt.Sprintf("%s", f.Usage)
}

func main() {
	log.SetPrefix("command-generator ")
	flag.Parse()
	args := flag.Args()
	if flag.NArg() != 2 {
		log.Fatalln("use: command-generator.go /path/to/commands.yml output-file.go")
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	commandPath := args[0]
	commandOut := filepath.Join(wd, args[1])

	log.Printf("building %s from %s\n", commandOut, commandPath)

	buildCommandFile(
		commandPath,
		commandOut,
		readCommandFile(commandPath),
	)
}

func readCommandFile(commandPath string) *CommandFile {
	contents, err := ioutil.ReadFile(commandPath)
	if err != nil {
		log.Fatalln("could not read contents", err)
	}

	cf := &CommandFile{Imports: map[string]string{
		"github.com/spf13/cobra": "",
	}}
	if err := yaml.Unmarshal(contents, cf); err != nil {
		log.Fatalf("failed to unmarshal YAML: %s", err)
	}

	return cf
}

func buildCommandFile(commandPath, commandOut string, cf *CommandFile) {
	headerTemplate := template.Must(template.New("").Funcs(generator.TemplateFuncs()).Parse(`
{{genlicense}}

{{gennotice "command-generator.go"}}

package {{.Package}}

{{genimports .Imports}}
`))

	typeTemplate := template.Must(template.New("").Parse(`
type {{.Unexported}} struct{
  {{ range $idx, $flag := .Flags}}// {{$flag.Unexported}} stores the value for "--{{$flag.Long}}"
  {{$flag.Unexported}} {{$flag.Type}}
  // {{$flag.Unexported}}IsSet is set to true if the user sets the flag
  {{$flag.Unexported}}IsSet bool
  {{ end }}
}
{{ range $idx, $flag := .Flags}}// {{$flag.Exported}} returns the value for "--{{$flag.Long}}" and if the user set it.
func (f *{{$.Unexported}}) {{$flag.Exported}}() ({{$flag.Type}}, bool) {
  return f.{{$flag.Unexported}}, f.{{$flag.Unexported}}IsSet
}
{{ end }}
`))

	newCommandTemplate := template.Must(template.New("").Parse(`
func new{{.Exported}}(run func(x {{.Unexported}}, cmd *cobra.Command, args []string) error) *cobra.Command {
  x := {{.Unexported}}{}
  cmd := &cobra.Command{
    Use: {{ printf "%q" .Use }},
    Short: {{ printf "%q" .Short }},
    Long: {{ printf "%q" .Long }},
    Example: {{ printf "%q" .IndentedExample }},
	Aliases: {{ .AliasStringArray }},
	Args: cobra.RangeArgs({{.ArgNRange}}),
  }
  cmd.PreRun = func(cmd *cobra.Command, args []string) {
    {{ range $idx, $flag := .Flags}}// {{$flag.Exported}} returns the value for "--{{$flag.Long}}" and if the user set it.

    x.{{$flag.Unexported}}IsSet = cmd.Flags().Changed("{{$flag.Long}}")
    {{ end }}
  }
  cmd.RunE = func(cmd *cobra.Command, args []string) error {
    return run(x, cmd, args)
  }

{{ range $idx, $flag := .Flags}}// {{$flag.Exported}} returns the value for "--{{$flag.Long}}" and if the user set it.
  cmd.Flags().{{$flag.CobraFuncName}}(
    &x.{{$flag.Unexported}},
	"{{$flag.Long}}",
	"{{$flag.Shorthand}}",
	{{$flag.DefaultValue}},
	"{{$flag.Usage}}",
  )
{{ end }}

  return cmd
}
`))

	manifestTypeTemplate := template.Must(template.New("").Parse(`
{{ if .IncludeManifest }}
type Manifest{{.Exported}} struct{
  {{ range $idx, $flag := .Flags}}// {{$flag.Unexported}} stores the value for "{{$flag.Exported}}"
  {{$flag.Exported}} {{$flag.Type}} {{$flag.ManifestStructTag}}
  {{ end }}
}
{{ end }}
`))

	buf := &bytes.Buffer{}

	// Write headers
	if err := headerTemplate.Execute(buf, cf); err != nil {
		log.Fatalln(err)
	}
	buf.WriteString("\n")

	// Write templates for each command
	for _, command := range cf.Commands {
		templs := []*template.Template{
			typeTemplate,
			newCommandTemplate,
			manifestTypeTemplate,
		}

		for _, templ := range templs {
			if err := templ.Execute(buf, command); err != nil {
				log.Fatalln(err)
			}
			buf.WriteString("\n")
		}
		writeMarkdown(commandPath, command)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatalf("failed to format code: %s", err)
	}

	f, err := os.Create(commandOut)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	if _, err := f.Write(formatted); err != nil {
		log.Fatalf("failed to write to %s: %s", commandOut, err)
	}
}

func camelCase(flagName string) string {
	flagNameSpaces := strings.ReplaceAll(flagName, "-", " ")
	flagNameTitled := strings.Title(flagNameSpaces)
	return strings.ReplaceAll(flagNameTitled, " ", "")
}

func writeMarkdown(commandPath string, cmd Command) {
	if cmd.MarkdownPath == "" {
		return
	}

	if !filepath.IsAbs(cmd.MarkdownPath) {
		cmd.MarkdownPath = filepath.Join(
			filepath.Dir(commandPath),
			cmd.MarkdownPath,
		)
	}

	headerTemplate := template.Must(template.New("").Funcs(generator.TemplateFuncs()).Parse(`
---
title: "{{ .Name }}"
linkTitle: "{{ .Name }}"
weight: 10
---

### Usage
kf {{ .Use }}

### Description
{{- if ne .Long "" }}
{{ .Long }}
{{ else if ne .Short "" }}
{{ .Short }}
{{ else }}
_No Description_
{{ end }}
{{ if (len .Aliases) gt 0 -}}
### Aliases
{{ range $idx, $alias := .Aliases }}
* {{ $alias }}
{{ end }}
{{ end }}
{{- if (len .Examples) gt 0 -}}
### Examples
{{ range $idx, $example:= .Examples }}
  {{ $example }}
{{ end }}
{{ end }}
{{- if (len .Args) gt 0 -}}
### Positional Arguments
{{ range $idx, $arg := .Args }}
##### {{ $arg.Name }}
{{ if $arg.Optional }}{{printf "%s (OPTIONAL)" $arg.Description}}{{else}}{{printf "%s (REQUIRED)" $arg.Description}}{{end}}
{{ end }}
{{ end }}
{{- if (len .Flags) gt 0 -}}
### Flags
{{ range $idx, $flag := .Flags }}
{{ $flag.MarkdownTitle }}
{{ $flag.MarkdownDefault }}
{{ end }}
{{ end }}
`))

	buf := &bytes.Buffer{}

	// Write templates for each command
	templs := []*template.Template{
		headerTemplate,
	}

	for _, templ := range templs {
		if err := templ.Execute(buf, cmd); err != nil {
			log.Fatalln(err)
		}
		buf.WriteString("\n")
	}

	if err := os.MkdirAll(filepath.Dir(cmd.MarkdownPath), 0777|os.ModeDir); err != nil {
		log.Fatalln(err)
	}

	f, err := os.Create(cmd.MarkdownPath)
	if err != nil {
		log.Fatalf("failed to create %s: %s", cmd.MarkdownPath, err)
	}
	defer f.Close()
	if _, err := f.Write(buf.Bytes()); err != nil {
		log.Fatalf("failed to write to %s: %s", cmd.MarkdownPath, err)
	}
}
