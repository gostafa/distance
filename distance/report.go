package distance

import "github.com/gostafa/distance/internal/shared/metrics"

// SchemaVersion is the version of the report schema produced by Analyze.
// Version 6 reports abstractness and instability next to distance. Version 5
// dropped size counts and declaration lists; packages keep path, coupling
// (Ca/Ce), and metrics.
const SchemaVersion = "6"

// ToolName is the canonical tool name embedded in reports.
const ToolName = "distance"

// MetricResult aliases the metrics package's result type; see its
// documentation for the applicability contract.
type MetricResult = metrics.MetricResult

// ToolInfo identifies the tool that produced a report.
type ToolInfo struct {
	// Name is the tool name embedded in reports; equals ToolName for this build.
	Name string
	// Version is the tool version string at analysis time.
	Version string
}

// Report is the complete, deterministic result of one analysis run.
// Packages are sorted by import path; ordering never depends on map
// iteration.
type Report struct {
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

// PackageReport carries one package's coupling facts and its metrics.
// Metrics appear in the fixed metric order: abstractness, instability,
// distance.
type PackageReport struct {
	// Path is the package's import path.
	Path string
	// Afferent counts analyzed packages importing this package (Ca).
	Afferent int
	// Efferent counts this package's in-scope imports (Ce), honoring the
	// configured dependency scope.
	Efferent int
	// Metrics holds the package-level metric results in the fixed metric
	// order: abstractness, instability, distance.
	Metrics []MetricResult
}
