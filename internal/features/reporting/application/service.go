package application

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"

	"github.com/gostafa/distance/distance"
	"github.com/gostafa/distance/internal/features/reporting/domain"
	"github.com/gostafa/distance/internal/features/reporting/ports/outbound"
	"github.com/gostafa/distance/internal/shared/metrics"
)

// jsonMarshal is a seam so tests can force encoding failures.
var jsonMarshal = json.Marshal

// Write renders the report in the given format into the sink. Options are
// read only by the text format.
func Write(
	report distance.Report,
	format domain.Format,
	sink outbound.Sink,
	opts domain.TextOptions,
) error {
	w, err := sink.Open()
	if err != nil {
		return err
	}

	renderErr := render(w, report, format, opts)
	if renderErr != nil {
		_ = w.Close()

		return renderErr
	}

	return w.Close()
}

func render(
	w io.Writer,
	report distance.Report,
	format domain.Format,
	opts domain.TextOptions,
) error {
	switch format {
	case domain.FormatText:
		_, err := io.WriteString(w, domain.Text(report, opts))

		return err
	case domain.FormatJSON:
		return renderJSON(w, report)
	case domain.FormatCSV:
		// WriteAll flushes; a separate header Write cannot surface bufio
		// errors until Flush, so header and records go through one call.
		rows := append([][]string{domain.CSVHeader()}, domain.CSVRecords(report)...)

		return csv.NewWriter(w).WriteAll(rows)
	case domain.FormatWeb:
		return renderWeb(w, report)
	default:
		return fmt.Errorf("unknown report format %q", format)
	}
}

// jsonReport mirrors the versioned report schema (§ output). Metric maps
// are orderedMetrics so keys always appear in the fixed metric order.
type jsonReport struct {
	// SchemaVersion is the report schema version.
	SchemaVersion string `json:"schema_version"`
	// Tool identifies the producing tool.
	Tool jsonTool `json:"tool"`
	// Packages are the analyzed packages in report order.
	Packages []jsonPackage `json:"packages"`
}

// String summarizes the report envelope for debugging.
func (r jsonReport) String() string {
	return fmt.Sprintf("schema %s, tool %v, %d packages", r.SchemaVersion, r.Tool, len(r.Packages))
}

type jsonTool struct {
	// Name is the tool's canonical name.
	Name string `json:"name"`
	// Version is the tool's build version.
	Version string `json:"version"`
}

type jsonPackage struct {
	// Path is the package's import path.
	Path string `json:"path"`
	// Afferent counts analyzed packages importing this package (Ca).
	Afferent int `json:"afferent"`
	// Efferent counts this package's in-scope imports (Ce).
	Efferent int `json:"efferent"`
	// Metrics maps metric names to results in the fixed order.
	Metrics orderedMetrics `json:"metrics"`
}

// String summarizes one package entry for debugging.
func (p jsonPackage) String() string {
	return fmt.Sprintf("%s: %d metrics", p.Path, len(p.Metrics))
}

// jsonMetric serializes one MetricResult. A non-applicable metric carries
// its reason and no value — never a fake zero.
type jsonMetric struct {
	// Scope is the kind of entity the metric describes.
	Scope string `json:"scope"`
	// Value is the metric value, present only when applicable.
	Value *float64 `json:"value,omitempty"`
	// Applicable reports whether the value may be read.
	Applicable bool `json:"applicable"`
	// Reason explains non-applicability or dropped components.
	Reason string `json:"reason,omitempty"`
	// Definition is the versioned formula identifier.
	Definition string `json:"definition"`
}

// orderedMetrics marshals as a JSON object keyed by metric name, preserving
// slice order (the fixed metric order).
type orderedMetrics []metrics.MetricResult

// MarshalJSON writes the object with keys in the fixed metric order.
func (m orderedMetrics) MarshalJSON() ([]byte, error) {
	return encodeOrderedMetrics(m)
}

// encodeOrderedMetrics assembles the ordered JSON object one name→metric
// pair at a time.
func encodeOrderedMetrics(results []metrics.MetricResult) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')

	for i, r := range results {
		if i > 0 {
			buf.WriteByte(',')
		}

		err := encodeMetricEntry(&buf, r)
		if err != nil {
			return nil, err
		}
	}

	buf.WriteByte('}')

	return buf.Bytes(), nil
}

// encodeMetricEntry writes one name→metric pair. A non-applicable metric
// carries its reason and no value — never a fake zero.
func encodeMetricEntry(buf *bytes.Buffer, r metrics.MetricResult) error {
	key, err := jsonMarshal(r.Name)
	if err != nil {
		return err
	}

	buf.Write(key)
	buf.WriteByte(':')

	out := jsonMetric{
		Scope:      string(r.Scope),
		Applicable: r.Applicable,
		Reason:     r.Reason,
		Definition: r.Definition,
	}
	if r.Applicable {
		value := r.Value
		out.Value = &value
	}

	encoded, err := jsonMarshal(out)
	if err != nil {
		return err
	}

	buf.Write(encoded)

	return nil
}

// buildJSONReport maps the report onto the versioned JSON schema. It is
// shared by the JSON format and the web report's embedded payload.
func buildJSONReport(report distance.Report) jsonReport {
	out := jsonReport{
		SchemaVersion: report.SchemaVersion,
		Tool:          jsonTool{Name: report.Tool.Name, Version: report.Tool.Version},
		Packages:      make([]jsonPackage, len(report.Packages)),
	}
	for i, pkg := range report.Packages {
		out.Packages[i] = jsonPackage{
			Path:     pkg.Path,
			Afferent: pkg.Afferent,
			Efferent: pkg.Efferent,
			Metrics:  orderedMetrics(pkg.Metrics),
		}
	}

	return out
}

func renderJSON(w io.Writer, report distance.Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	return enc.Encode(buildJSONReport(report))
}
