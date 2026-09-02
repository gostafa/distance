// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package distance

import (
	"github.com/gostafa/distance/internal/shared/metrics"
)

type (

	// reporter exposes the module path and package list of a Report.
	reporter interface {
		ModulePath() string
		PackageList() []PackageReport
	}

	// configured exposes the pattern list and dependency scope of a Config.
	configured interface {
		PatternList() []string
		ScopeName() string
	}

	// Config controls a single Analyze run.
	Config struct {
		Directory        string
		DependencyScope  DependencyScope
		Patterns         []string
		BuildTags        []string
		Workers          int
		IncludeTests     bool
		IncludeGenerated bool
		ContinueOnError  bool
	}

	// PackageReport holds metrics and coupling for one analyzed package.
	PackageReport struct {
		Path     string
		Metrics  []metrics.MetricResult
		Afferent int
		Efferent int
	}

	// Report is the complete analysis result for one configuration.
	Report struct {
		// SchemaVersion identifies the report schema; it equals the
		// SchemaVersion constant for reports this build produces.
		SchemaVersion string
		// Tool records the tool name and version that produced the report.
		Tool ToolInfo
		// Module is the analyzed main module's path, when known. Renderers
		// use it to show package paths relative to the module root.
		Module string
		// Packages holds one entry per analyzed package, sorted by import path.
		Packages []PackageReport
	}

	// ToolInfo names the binary that produced a Report.
	ToolInfo struct {
		// Name is the tool name embedded in reports.
		Name string
		// Version is the tool version string at analysis time.
		Version string
	}

	configError struct {
		message string
	}
)
