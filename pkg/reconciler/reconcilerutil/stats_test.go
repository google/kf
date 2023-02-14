package reconcilerutil

import (
	"bytes"
	"testing"
	"time"

	"github.com/google/kf/v2/pkg/kf/testutil"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/types"
)

func newFakeClock() *fakeClock {
	return &fakeClock{
		time: time.Unix(0, 0),
	}
}

type fakeClock struct {
	time time.Time
}

// Now returns the current internal time of the fake clock.
func (fc *fakeClock) Now() time.Time {
	return fc.time
}

// AdvanceReport advances to the next reporting time.
func (fc *fakeClock) AdvanceReport() {
	fc.time = fc.time.Add(reportFrequency)
}

func TestStructuredStatsReporter_now(t *testing.T) {
	t.Parallel()

	t.Run("defaults to system clock", func(t *testing.T) {
		s := &StructuredStatsReporter{
			Clock: nil,
		}

		var times []time.Time

		// Alternate time sources
		for i := 0; i < 5; i++ {
			times = append(times, time.Now())
			time.Sleep(50 * time.Millisecond)

			times = append(times, s.now())
			time.Sleep(50 * time.Millisecond)
		}

		// Expect time to be strictly increasing.
		for i := 1; i < len(times); i++ {
			if times[i-1].Compare(times[i]) >= 0 {
				t.Fatalf(
					"Times aren't increasing time %d (%q) was after %d (%q)",
					i-1, times[i-1],
					i, times[i])
			}
		}
	})

	t.Run("clock can be overridden", func(t *testing.T) {
		var count int64
		s := &StructuredStatsReporter{
			Clock: func() time.Time {
				count++
				return time.Unix(count, 0)
			},
		}

		for i := int64(1); i < 100; i++ {
			nowUnix := s.now().Unix()
			if nowUnix != i {
				t.Fatalf(
					"Expected time at step %d to be %d got %d",
					i-1,
					i,
					nowUnix,
				)
			}
		}
	})
}

type bytesBufferSyncer struct {
	bytes.Buffer
}

func (*bytesBufferSyncer) Sync() error {
	return nil
}

func scenarioOutput(scenario func(*StructuredStatsReporter, *fakeClock)) (report string, generated bool) {
	// Initialize
	clock := newFakeClock()
	buf := &bytesBufferSyncer{}
	encoderCfg := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		NameKey:        "logger",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), buf, zap.DebugLevel)
	sugaredLogger := zap.New(core).WithOptions().Sugar()
	reporter := &StructuredStatsReporter{
		Logger: sugaredLogger,
		Clock:  clock.Now,
	}

	// Run scenario
	scenario(reporter, clock)

	// Get output
	clock.AdvanceReport()
	generated = reporter.MaybeReport()
	return string(buf.Bytes()), generated
}

func TestStructuredStatsReporter_ReportQueueDepth(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		scenario          func(*StructuredStatsReporter, *fakeClock)
		expectLogContains string
	}{
		"last queue depth wins": {
			scenario: func(reporter *StructuredStatsReporter, _ *fakeClock) {
				reporter.ReportQueueDepth(200)
				reporter.ReportQueueDepth(100)
				reporter.ReportQueueDepth(50)
			},
			expectLogContains: `"currentQueueDepth":50`,
		},
		"empty queue depth is zero": {
			scenario: func(reporter *StructuredStatsReporter, _ *fakeClock) {
				// No reports
			},
			expectLogContains: `"currentQueueDepth":0`,
		},
		"queue depth remains after report": {
			scenario: func(reporter *StructuredStatsReporter, clock *fakeClock) {
				log := reporter.Logger
				reporter.Logger = nil

				reporter.ReportQueueDepth(200)
				clock.AdvanceReport()
				reporter.MaybeReport()

				reporter.Logger = log
			},
			expectLogContains: `"currentQueueDepth":200`,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			output, generated := scenarioOutput(tc.scenario)
			testutil.AssertTrue(t, "report generated", generated)
			testutil.AssertContainsAll(t, output, []string{tc.expectLogContains})
		})
	}
}

func TestStructuredStatsReporter_ReportReconcile(t *testing.T) {
	t.Parallel()

	blankName := types.NamespacedName{}

	cases := map[string]struct {
		scenario        func(*StructuredStatsReporter, *fakeClock)
		expectedOutputs []string
	}{
		"2 reconciliations per second": {
			scenario: func(reporter *StructuredStatsReporter, _ *fakeClock) {
				for i := 0; i < (int(reportFrequency)/int(time.Second))*2; i++ {
					reporter.ReportReconcile(time.Second*5, "true", blankName)
				}
			},
			expectedOutputs: []string{
				`"reconciliationsPerSecond":2`,
			},
		},
		"mean reconciliation time reported": {
			scenario: func(reporter *StructuredStatsReporter, _ *fakeClock) {
				reporter.ReportReconcile(time.Second*0, "true", blankName)
				reporter.ReportReconcile(time.Second*10, "true", blankName)
			},
			expectedOutputs: []string{
				`"averageMSPerReconciliation":5000`,
			},
		},
		"type tallys are correct": {
			scenario: func(reporter *StructuredStatsReporter, _ *fakeClock) {
				for i := 0; i < 3; i++ {
					reporter.ReportReconcile(time.Second*0, "true", blankName)
				}
				for i := 0; i < 4; i++ {
					reporter.ReportReconcile(time.Second*0, "false", blankName)
				}
				for i := 0; i < 5; i++ {
					reporter.ReportReconcile(time.Second*0, "x", blankName)
				}
			},
			expectedOutputs: []string{
				`"success":3`,
				`"failure":4`,
				`"unknown":5`,
			},
		},
		"no reports": {
			scenario: func(reporter *StructuredStatsReporter, _ *fakeClock) {
				// No reports
			},
			expectedOutputs: []string{
				`"reconciliationsPerSecond":0`,
				`"averageMSPerReconciliation":0`,
				`"statuses":{"failure":0,"success":0,"unknown":0}`,
			},
		},
		"values reset after report": {
			scenario: func(reporter *StructuredStatsReporter, clock *fakeClock) {
				log := reporter.Logger
				reporter.Logger = nil

				reporter.ReportReconcile(time.Second*1, "true", blankName)
				reporter.ReportReconcile(time.Second*1, "false", blankName)
				reporter.ReportReconcile(time.Second*1, "x", blankName)

				clock.AdvanceReport()
				reporter.MaybeReport()

				reporter.Logger = log
			},
			expectedOutputs: []string{
				`"reconciliationsPerSecond":0`,
				`"averageMSPerReconciliation":0`,
				`"statuses":{"failure":0,"success":0,"unknown":0}`,
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			output, generated := scenarioOutput(tc.scenario)
			testutil.AssertTrue(t, "report generated", generated)
			testutil.AssertContainsAll(t, output, tc.expectedOutputs)
		})
	}
}

func TestStructuredStatsReporter_MaybeReport(t *testing.T) {
	t.Parallel()

	t.Run("nil logger still succeeds", func(t *testing.T) {
		reporter := &StructuredStatsReporter{
			Logger: nil,
		}

		resp := reporter.MaybeReport()
		testutil.AssertTrue(t, "response", resp)
	})

	t.Run("values reset after report", func(t *testing.T) {
		out, generated := scenarioOutput(func(reporter *StructuredStatsReporter, clock *fakeClock) {
			tmpLogger := reporter.Logger
			reporter.Logger = nil

			reporter.ReportQueueDepth(100)
			clock.AdvanceReport()
			reporter.MaybeReport()

			reporter.Logger = tmpLogger
		})
		testutil.AssertTrue(t, "report generated", generated)
		testutil.AssertContainsAll(t, out, []string{
			`"currentQueueDepth":100`,
			`"reconciliationsPerSecond":0`,
			`"averageMSPerReconciliation":0`,
			`"statuses":{"failure":0,"success":0,"unknown":0}`,
		})
	})

}
