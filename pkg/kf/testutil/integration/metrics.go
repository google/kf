// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package integration

import (
	"context"
	"errors"
	"log"
	"os"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/logging"
	mrpb "google.golang.org/genproto/googleapis/api/monitoredres"
)

const (
	labelsKey = "labels"
)

// Metric is used with Write() to emit metrics.
type Metric interface {
	toMap() map[string]interface{}
}

// MetricLabels are used to store labels on metrics.
type MetricLabels map[string]string

// Int64Gauge writes an int64 value along with a name.
type Int64Gauge struct {
	Name   string
	Units  string
	Labels MetricLabels
	Value  int64
}

func (g Int64Gauge) toMap() map[string]interface{} {
	return map[string]interface{}{
		"name":    g.Name,
		"value":   g.Value,
		labelsKey: g.Labels,
	}
}

// Float64Gauge writes an float64 value along with a name.
type Float64Gauge struct {
	Name   string
	Labels MetricLabels
	Value  float64
}

func (g Float64Gauge) toMap() map[string]interface{} {
	return map[string]interface{}{
		"name":    g.Name,
		"value":   g.Value,
		labelsKey: g.Labels,
	}
}

// ContextWithStackdriverOutput is a wrapper around ContextWithMetricLogger
// that uses a Stackdriver client. If setting up the client fails, then a
// writer isn't set and the error is logged to the standard logger.
func ContextWithStackdriverOutput(ctx context.Context) context.Context {
	if os.Getenv(EgressMetricsEnvVar) != "true" {
		return ctx
	}

	// Create a Client
	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		var err error
		projectID, err = metadata.ProjectID()
		if err != nil {
			log.Printf("failed to fetch GCP project ID: %v", err)
			return ctx
		}
	}

	client, err := logging.NewClient(ctx, projectID)
	if err != nil {
		log.Printf("failed to setup Stackdriver client: %v", err)
		return ctx
	}

	return ContextWithMetricLogger(ctx, client.Logger("kf-tests-metrics"))
}

type metricOutputType struct{}

// MetricLogger is used to write metrics. It is implemented by logging.Logger.
type MetricLogger interface {
	LogSync(ctx context.Context, e logging.Entry) error
}

// ContextWithMetricLogger returns a Context that stores the given
// MetricLogger.  The writer is used when writing metrics. It should be used
// with MetricLoggerFromContext().
func ContextWithMetricLogger(ctx context.Context, out MetricLogger) context.Context {
	return context.WithValue(ctx, metricOutputType{}, out)
}

// MetricLoggerFromContext returns a MetricLogger from the context that was
// set with ContextWithMetricLogger(). If one was never set, then a nil is
// returned.
func MetricLoggerFromContext(ctx context.Context) MetricLogger {
	w, ok := ctx.Value(metricOutputType{}).(MetricLogger)
	if !ok {
		return nil
	}
	return w
}

type metricLabelType struct{}

// ContextWithMetricLabels returns a Context that stores the given labels. The
// labels are used when writing a Metric. If there are labels already set on
// the context, they will be joined deferring to the new labels.
func ContextWithMetricLabels(ctx context.Context, labels MetricLabels) context.Context {
	return context.WithValue(
		ctx,
		metricLabelType{},
		joinLabels(MetricLabelsFromContext(ctx), labels),
	)
}

// MetricLabelsFromContext returns MetricLabels from the Context. If there
// aren't any MetricLabels set, an empty MetricLabels object is returned.
func MetricLabelsFromContext(ctx context.Context) MetricLabels {
	t, ok := ctx.Value(metricLabelType{}).(MetricLabels)
	if !ok {
		return nil
	}
	return t
}

// WriteMetric writes the metric as a structured log to the output set by
// ContextWithMetricLogger(). If no writer is set, then this function is a
// NOP.
//
// Any labels in the `context.Context` set by ContextWithMetricLabels()  are
// joined with the labels in the Metric.
func WriteMetric(ctx context.Context, m Metric) {
	out := MetricLoggerFromContext(ctx)
	if out == nil {
		return
	}

	// Skip writing metrics if the context is done
	if errors.Is(ctx.Err(), context.Canceled) {
		log.Println("skipping metric upload due to cancelled context")
		return
	}

	mm := m.toMap()

	// Move the labels to the logging.Entry.
	labels, ok := mm[labelsKey].(MetricLabels)
	if !ok {
		labels = nil
	}
	delete(mm, labelsKey)
	labels = joinLabels(MetricLabelsFromContext(ctx), labels)

	if err := out.LogSync(ctx, logging.Entry{
		Payload: mm,
		Labels:  labels,
		Resource: &mrpb.MonitoredResource{
			// A resource type used to indicate that a log is not associated
			// with any specific resource.
			Type: "global",
		},
	}); err != nil {
		log.Printf("failed to write log: %v", err)
		return
	}
}

// joinLabels returns a MetricLabels object with all the passed MetricLabels
// joined. The MetricLabels are applied in order, and therefore the earlier in
// the array, the less priority it gets.
func joinLabels(ts ...MetricLabels) MetricLabels {
	result := MetricLabels{}

	for _, m := range ts {
		for k, v := range m {
			result[k] = v
		}
	}

	return result
}
