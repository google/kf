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
	if err := TabbedWriter(os.Stdout, func(w io.Writer) error {
		if _, err := fmt.Fprintln(w, "OS\tAGE"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, "Linux\t20y"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, "DOS\t40y"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, "BeOS\t20y"); err != nil {
			return err
		}

		return nil
	}); err != nil {
		panic(err)
	}

	// Output: OS     AGE
	// Linux  20y
	// DOS    40y
	// BeOS   20y
}

func ExampleIndentWriter() {
	w := os.Stdout
	if _, err := fmt.Fprintln(w, "Level0"); err != nil {
		panic(err)
	}
	if err := IndentWriter(w, func(w io.Writer) error {
		if _, err := fmt.Fprintln(w, "Level1"); err != nil {
			return err
		}
		if err := IndentWriter(w, func(w io.Writer) error {
			if _, err := fmt.Fprintln(w, "Level2"); err != nil {
				return err
			}

			return nil
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		panic(err)
	}

	// Output: Level0
	//   Level1
	//     Level2
}

func ExampleSectionWriter_empty() {
	if err := SectionWriter(os.Stdout, "SectionName", func(_ io.Writer) error {
		// No output
		return nil
	}); err != nil {
		panic(err)
	}

	// Output: SectionName: <empty>
}

func ExampleSectionWriter_populated() {
	if err := SectionWriter(os.Stdout, "OperatingSystems", func(w io.Writer) error {
		if _, err := fmt.Fprintln(w, "Linux:\tOSS"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, "DOS:\tPaid"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, "BeOS:\tDead"); err != nil {
			return err
		}

		return nil
	}); err != nil {
		panic(err)
	}

	// Output: OperatingSystems:
	//   Linux:  OSS
	//   DOS:    Paid
	//   BeOS:   Dead
}
