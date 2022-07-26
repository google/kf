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

package logging

import (
	"bytes"
	"context"
	"io"

	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"knative.dev/pkg/logging"
)

// SetupLogger returns a context with a logger setup on it.
func SetupLogger(ctx context.Context, w io.Writer) context.Context {
	logger := zap.New(&loggerCore{w: w})

	return logging.WithLogger(ctx, logger.Sugar())
}

type loggerCore struct {
	w      io.Writer
	fields []zapcore.Field
}

var _ zapcore.Core = (*loggerCore)(nil)

func (c *loggerCore) Enabled(_ zapcore.Level) bool {
	return true
}

// With adds structured context to the Core.
func (c *loggerCore) With(fields []zapcore.Field) zapcore.Core {
	clone := c.clone()
	clone.fields = append(clone.fields, fields...)
	return clone
}

// Check determines whether the supplied Entry should be logged (using the
// embedded LevelEnabler and possibly some extra logic). If the entry
// should be logged, the Core adds itself to the CheckedEntry and returns
// the result.
//
// Callers must use Check before calling Write.
func (c *loggerCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return ce.AddCore(ent, c)
}

var (
	debugColor = color.New(color.FgHiBlue, color.Bold)
	warnColor  = color.New(color.FgHiYellow, color.Bold)
	errorColor = color.New(color.FgHiRed, color.Bold)
	fieldColor = color.New(color.FgGreen)
)

// Write serializes the Entry and any Fields supplied at the log site and
// writes them to their destination.
//
// If called, Write should always log the Entry and Fields; it should not
// replicate the logic of Check.
func (c *loggerCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	var sb bytes.Buffer
	message := ent.Message

	// Add the log level prefix.
	switch ent.Level {
	case zapcore.InfoLevel:
		// NOP
	case zapcore.DebugLevel:
		sb.WriteString(debugColor.Sprint(ent.Level.CapitalString()) + " ")
	case zapcore.WarnLevel:
		sb.WriteString(warnColor.Sprint(ent.Level.CapitalString()) + " ")
	case zapcore.ErrorLevel, zapcore.DPanicLevel, zapcore.PanicLevel:
		sb.WriteString(errorColor.Sprint(ent.Level.CapitalString()) + " ")
	}

	// Add the fields as [fieldA] [fieldB]
	for _, field := range c.fields {
		if field.String == "" {
			sb.WriteString(fieldColor.Sprintf("[%s] ", field.Key))
		} else {
			sb.WriteString(fieldColor.Sprintf("[%s/%s] ", field.Key, field.String))
		}
	}

	sb.WriteString(message)
	sb.WriteString("\n")

	if _, err := c.w.Write(sb.Bytes()); err != nil {
		return err
	}

	return nil
}

// Sync flushes buffered logs (if any).
func (c *loggerCore) Sync() error {
	return nil
}

func (c *loggerCore) clone() *loggerCore {
	return &loggerCore{
		w:      c.w,
		fields: c.fields,
	}
}
