// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package formats_test

import (
	"bytes"
	"fmt"
	"testing"

	reporting "github.com/gostafa/distance/internal/features/reporting/adapters/formats"
	domain "github.com/gostafa/distance/internal/features/reporting/adapters/presentation"
	reportmodel "github.com/gostafa/distance/internal/features/reporting/domain"
)

const (
	benchPackages = 200
	benchSpread   = 100.0
)

func benchReport() *reportmodel.Report {
	report := &reportmodel.Report{
		SchemaVersion: "6",
		ToolName:      "distance",
		ToolVersion:   "v0.0.0-bench",
		Module:        "example.com/m",
		Packages:      make([]reportmodel.PackageReport, benchPackages),
	}

	for index := range report.Packages {
		value := float64(index%100) / benchSpread

		report.Packages[index] = reportmodel.PackageReport{
			Path:     fmt.Sprintf("example.com/m/internal/pkg%03d", index),
			Afferent: index % 7,
			Efferent: index % 11,
			Metrics: []reportmodel.MetricResult{
				{
					Name:       reportmodel.MetricAbstractness,
					Scope:      "package",
					Value:      value,
					Applicable: true,
					Definition: "a",
				},
				{
					Name:       reportmodel.MetricInstability,
					Scope:      "package",
					Value:      1 - value,
					Applicable: true,
					Definition: "i",
				},
				{
					Name:       reportmodel.MetricDistance,
					Scope:      "package",
					Value:      value / 2,
					Applicable: true,
					Definition: "d",
				},
			},
		}
	}

	return report
}

func BenchmarkWrite(b *testing.B) {
	report := benchReport()

	for _, format := range []domain.Format{domain.FormatText, domain.FormatJSON, domain.FormatCSV, domain.FormatWeb} {
		b.Run(format, func(b *testing.B) {
			sink := &bytes.Buffer{}
			opts := &reporting.WriteOptions{Format: format}

			b.ReportAllocs()

			for b.Loop() {
				sink.Reset()

				err := reporting.Write(report, bufWriteCloser{sink}, opts)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
