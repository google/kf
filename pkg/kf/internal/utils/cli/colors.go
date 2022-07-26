// Copyright 2019 Google LLC

package utils

import (
	"github.com/fatih/color"
)

var (
	// WarningColor is the color to use for warnings.
	WarningColor = color.New(color.FgHiYellow, color.Bold)
	// Warnf behaves like Sprintf for a warning color.
	Warnf = WarningColor.Sprintf
	// Heading is the color for headings.
	Heading = color.New(color.FgWhite, color.Bold)
	// Muted is the color for headings you want muted.
	Muted = color.New(color.FgWhite, color.Bold)
	// TipColor is the color to show for tips.
	TipColor = color.New(color.FgGreen, color.Bold)
)
