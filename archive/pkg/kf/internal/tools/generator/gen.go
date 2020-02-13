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

package generator

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"
	"time"
)

var licenseHeader = template.Must(template.New("").Parse(`// Copyright {{.year}} Google LLC
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
// limitations under the License.`))

// GenLicense generates the license header.
func GenLicense() string {
	buf := &bytes.Buffer{}
	licenseHeader.Execute(buf, map[string]interface{}{
		"year": time.Now().Year(),
	})
	return buf.String()
}

// GenImports generates an import declaration from the given map of import:alias
// pairs. If an alias is blank, then the path is imported directly.
func GenImports(pathAlias map[string]string) string {
	if len(pathAlias) == 0 {
		return ""
	}

	imports := []string{}
	for path, alias := range pathAlias {
		if alias != "" {
			alias += " "
		}

		imports = append(imports, fmt.Sprintf("\t%s%q", alias, path))
	}

	sort.Strings(imports)

	out := &bytes.Buffer{}
	fmt.Fprintln(out, "import (")
	for _, i := range imports {
		fmt.Fprintln(out, i)
	}
	fmt.Fprintln(out, ")")

	return out.String()
}

// GenNotice generates a notice that the file was auto-generated.
func GenNotice(file string) string {
	return fmt.Sprintf("// This file was generated with %s, DO NOT EDIT IT.", file)
}

// TemplateFuncs returns the functions provided by generator available for
// consumption in a template.
func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"genlicense": GenLicense,
		"genimports": GenImports,
		"gennotice":  GenNotice,
		"title":      strings.Title,
		"lower":      strings.ToLower,
	}
}
