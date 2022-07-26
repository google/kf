// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package doctor

import (
	"context"
	"fmt"
	"testing"
)

func ExampleDiagnostic_Run() {
	d := NewDefaultDiagnostic()

	d.Run(context.Background(), "sub-test-good", func(ctx context.Context, d *Diagnostic) {
		d.Log("no output")
	})

	d.Run(context.Background(), "sub-test-warn", func(ctx context.Context, d *Diagnostic) {
		d.Warn("warning output")
	})

	d.Run(context.Background(), "sub-test-bad", func(ctx context.Context, d *Diagnostic) {
		d.Error("output")
	})

	// Output: === RUN	doctor/sub-test-good
	// --- PASS: doctor/sub-test-good
	// === RUN	doctor/sub-test-warn
	// === LOG	doctor/sub-test-warn
	// warning output
	//
	// --- WARN: doctor/sub-test-warn
	// === RUN	doctor/sub-test-bad
	// === LOG	doctor/sub-test-bad
	// output
	//
	// --- FAIL: doctor/sub-test-bad
}

func ExampleDiagnostic_GatedRun() {
	d := NewDefaultDiagnostic()

	d.GatedRun(context.Background(), "validate-db-creds", func(ctx context.Context, d *Diagnostic) {
		d.Error("bad creds")
	})

	d.GatedRun(context.Background(), "validate-db-version", func(ctx context.Context, d *Diagnostic) {
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

	d.Run(context.Background(), "example", func(ctx context.Context, d *Diagnostic) {
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

	d.Run(context.Background(), "example", func(ctx context.Context, d *Diagnostic) {
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

	d.Run(context.Background(), "example", func(ctx context.Context, d *Diagnostic) {
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

	d.Run(context.Background(), "example", func(ctx context.Context, d *Diagnostic) {
		d.Logf("only printed on failure %d", 42)
	})

	// Output: === RUN	doctor/example
	// --- PASS: doctor/example
}

func ExampleDiagnostic_Warnf() {
	d := NewDefaultDiagnostic()

	d.Run(context.Background(), "example", func(ctx context.Context, d *Diagnostic) {
		d.Warnf("%d", 42)
		d.Log("after")
	})

	// Output: === RUN	doctor/example
	// === LOG	doctor/example
	// 42
	// after
	// --- WARN: doctor/example
}

func ExampleDiagnostic_Report() {
	d := NewDefaultDiagnostic()

	d.Run(context.Background(), "good", func(ctx context.Context, d *Diagnostic) {
		d.Logf("only printed on failure %d", 42)
	})

	d.Run(context.Background(), "bad", func(ctx context.Context, d *Diagnostic) {
		d.Run(context.Background(), "no-issue", func(ctx context.Context, d *Diagnostic) {})
		d.Run(context.Background(), "fails", func(ctx context.Context, d *Diagnostic) {
			d.Fail()
		})
	})

	fmt.Println("Final Report:")
	d.Report()

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

func TestStatusInvariant(t *testing.T) {
	var emptyStatus diagnosticStatus
	if statusPass != emptyStatus {
		t.Error("expected empty status to be statusPass")
	}

	// XXX: For readability, don't refactor these expressions;
	// they should be in strictly increasing order to assert the
	// transitive property.

	if !(statusWarn > statusPass) {
		t.Errorf("expected warn %v > pass %v", statusWarn, statusPass)
	}

	if !(statusFail > statusWarn) {
		t.Errorf("expected fail %v > warn %v", statusFail, statusWarn)
	}
}
