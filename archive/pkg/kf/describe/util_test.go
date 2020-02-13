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

package describe

import (
	"fmt"
	"io"
	"os"
)

func ExampleTabbedWriter() {
	TabbedWriter(os.Stdout, func(w io.Writer) {
		fmt.Fprintln(w, "OS\tAGE")
		fmt.Fprintln(w, "Linux\t20y")
		fmt.Fprintln(w, "DOS\t40y")
		fmt.Fprintln(w, "BeOS\t20y")
	})

	// Output: OS     AGE
	// Linux  20y
	// DOS    40y
	// BeOS   20y
}

func ExampleIndentWriter() {
	w := os.Stdout
	fmt.Fprintln(w, "Level0")
	IndentWriter(w, func(w io.Writer) {
		fmt.Fprintln(w, "Level1")
		IndentWriter(w, func(w io.Writer) {
			fmt.Fprintln(w, "Level2")
		})
	})

	// Output: Level0
	//   Level1
	//     Level2
}

func ExampleSectionWriter_empty() {
	SectionWriter(os.Stdout, "SectionName", func(_ io.Writer) {
		// No output
	})

	// Output: SectionName: <empty>
}

func ExampleSectionWriter_populated() {
	SectionWriter(os.Stdout, "OperatingSystems", func(w io.Writer) {
		fmt.Fprintln(w, "Linux:\tOSS")
		fmt.Fprintln(w, "DOS:\tPaid")
		fmt.Fprintln(w, "BeOS:\tDead")
	})

	// Output: OperatingSystems:
	//   Linux:  OSS
	//   DOS:    Paid
	//   BeOS:   Dead
}
