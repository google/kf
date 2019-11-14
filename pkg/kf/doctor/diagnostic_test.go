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
package doctor

import "fmt"

func ExampleDiagnostic_Run() {
	d := NewDefaultDiagnostic()

	d.Run("sub-test-good", func(d *Diagnostic) {
		d.Log("no output")
	})

	d.Run("sub-test-bad", func(d *Diagnostic) {
		d.Error("output")
	})

	// Output: === RUN	doctor/sub-test-good
	// --- PASS: doctor/sub-test-good
	// === RUN	doctor/sub-test-bad
	// === LOG	doctor/sub-test-bad
	// output
	//
	// --- FAIL: doctor/sub-test-bad
}

func ExampleDiagnostic_GatedRun() {
	d := NewDefaultDiagnostic()

	d.GatedRun("validate-db-creds", func(d *Diagnostic) {
		d.Error("bad creds")
	})

	d.GatedRun("validate-db-version", func(d *Diagnostic) {
		// never reached because the previous run failed
	})

	// Output: === RUN	doctor/validate-db-creds
	// === LOG	doctor/validate-db-creds
	// bad creds
	//
	// --- FAIL: doctor/validate-db-creds
}

func ExampleDiagnostic_Fatal() {
	d := NewDefaultDiagnostic()

	d.Run("example", func(d *Diagnostic) {
		d.Log("before")
		d.Fatal("immediate exit")
		d.Log("after")
	})

	// Output: === RUN	doctor/example
	// === LOG	doctor/example
	// before
	// immediate exit
	// --- FAIL: doctor/example
}

func ExampleDiagnostic_Fatalf() {
	d := NewDefaultDiagnostic()

	d.Run("example", func(d *Diagnostic) {
		d.Fatalf("%d", 42)
		d.Log("after")
	})

	// Output: === RUN	doctor/example
	// === LOG	doctor/example
	// 42
	// --- FAIL: doctor/example
}

func ExampleDiagnostic_Errorf() {
	d := NewDefaultDiagnostic()

	d.Run("example", func(d *Diagnostic) {
		d.Errorf("%d", 42)
		d.Log("after")
	})

	// Output: === RUN	doctor/example
	// === LOG	doctor/example
	// 42
	// after
	// --- FAIL: doctor/example
}

func ExampleDiagnostic_Logf() {
	d := NewDefaultDiagnostic()

	d.Run("example", func(d *Diagnostic) {
		d.Logf("only printed on failure %d", 42)
	})

	// Output: === RUN	doctor/example
	// --- PASS: doctor/example
}

func ExampleDiagnostic_Report() {
	d := NewDefaultDiagnostic()

	d.Run("good", func(d *Diagnostic) {
		d.Logf("only printed on failure %d", 42)
	})

	d.Run("bad", func(d *Diagnostic) {
		d.Run("no-issue", func(d *Diagnostic) {})
		d.Run("fails", func(d *Diagnostic) {
			d.Fail()
		})
	})

	if _, err := fmt.Println("Final Report:"); err != nil {
		panic(err)
	}
	if err := d.Report(); err != nil {
		panic(err)
	}

	// Output: === RUN	doctor/good
	// --- PASS: doctor/good
	// === RUN	doctor/bad
	// === RUN	doctor/bad/no-issue
	// === RUN	doctor/bad/fails
	// === LOG	doctor/bad/fails
	// === LOG	doctor/bad
	// --- FAIL: doctor/bad
	//     --- PASS: doctor/bad/no-issue
	//     --- FAIL: doctor/bad/fails
	// Final Report:
	// --- FAIL: doctor
	//     --- PASS: doctor/good
	//     --- FAIL: doctor/bad
	//         --- PASS: doctor/bad/no-issue
	//         --- FAIL: doctor/bad/fails
}
