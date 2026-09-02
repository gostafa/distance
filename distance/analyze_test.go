package distance_test

import (
	"context"
	"math"
	"reflect"
	"sync"
	"testing"

	"github.com/gostafa/distance/distance"
)

const epsilon = 1e-12

// The default-config report is loaded once and shared read-only across the
// many default-config tests — package loading dominates test time, so this
// avoids re-running the analyzer for every case. Config-varying tests (mutate
// != nil) still load fresh.
var (
	defaultOnce   sync.Once
	defaultReport distance.Report
	defaultErr    error
)

func analyzeFixture(t *testing.T, mutate func(*distance.Config)) distance.Report {
	t.Helper()

	if mutate == nil {
		defaultOnce.Do(func() {
			defaultReport, defaultErr = distance.Analyze(
				context.Background(), distance.Config{Directory: "../testdata/fixture"},
			)
		})

		if defaultErr != nil {
			t.Fatal(defaultErr)
		}

		return defaultReport
	}

	cfg := distance.Config{Directory: "../testdata/fixture"}
	mutate(&cfg)

	report, err := distance.Analyze(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}

	return report
}

func findPackage(t *testing.T, report distance.Report, path string) distance.PackageReport {
	t.Helper()

	for _, pkg := range report.Packages {
		if pkg.Path == path {
			return pkg
		}
	}

	t.Fatalf("package %s not in report", path)

	return distance.PackageReport{}
}

func findType(t *testing.T, pkg distance.PackageReport, name string) distance.TypeReport {
	t.Helper()

	for _, typ := range pkg.Types {
		if typ.Name == name {
			return typ
		}
	}

	t.Fatalf("type %s not in package %s", name, pkg.Path)

	return distance.TypeReport{}
}

func metric(
	t *testing.T,
	results []distance.MetricResult,
	name string,
) distance.MetricResult {
	t.Helper()

	for _, r := range results {
		if r.Name == name {
			return r
		}
	}

	t.Fatalf("metric %s not present in %v", name, results)

	return distance.MetricResult{}
}

func wantValue(t *testing.T, results []distance.MetricResult, name string, want float64) {
	t.Helper()

	r := metric(t, results, name)
	if !r.Applicable {
		t.Fatalf("%s not applicable (%s), want %v", name, r.Reason, want)
	}

	if math.Abs(r.Value-want) > epsilon {
		t.Fatalf("%s = %v, want %v", name, r.Value, want)
	}
}

func wantNotApplicable(t *testing.T, results []distance.MetricResult, name string) {
	t.Helper()

	r := metric(t, results, name)
	if r.Applicable {
		t.Fatalf("%s applicable with value %v, want n/a", name, r.Value)
	}

	if r.Reason == "" {
		t.Fatalf("%s n/a without reason", name)
	}
}

func TestAnalyzeFixtureOrdering(t *testing.T) {
	report := analyzeFixture(t, nil)

	wantOrder := []string{
		"example.com/fixture/embedding",
		"example.com/fixture/gen",
		"example.com/fixture/generics",
		"example.com/fixture/isolated",
		"example.com/fixture/multifile",
		"example.com/fixture/orders",
		"example.com/fixture/store",
	}
	if len(report.Packages) != len(wantOrder) {
		t.Fatalf("got %d packages", len(report.Packages))
	}

	for i, path := range wantOrder {
		if report.Packages[i].Path != path {
			t.Fatalf("packages[%d] = %s, want %s", i, report.Packages[i].Path, path)
		}
	}

	if report.SchemaVersion != distance.SchemaVersion ||
		report.Tool.Name != distance.ToolName {
		t.Fatalf("report header = %+v", report)
	}
}

func TestAnalyzeOrderType(t *testing.T) {
	report := analyzeFixture(t, nil)
	order := findType(t, findPackage(t, report, "example.com/fixture/orders"), "Order")

	if order.Kind != "struct" || !order.Exported || order.Position.File != "orders/orders.go" {
		t.Fatalf("order type details = %+v", order)
	}
	if len(order.FieldDetails) != 3 || order.FieldDetails[0].Name != "ID" {
		t.Fatalf("order field details = %+v", order.FieldDetails)
	}
	if len(order.MethodDetails) != 3 {
		t.Fatalf("order method details = %+v", order.MethodDetails)
	}
	grandTotal := findFunctionReport(t, order.MethodDetails, "GrandTotal")
	if grandTotal.Receiver != "Order" || grandTotal.Lines != 6 {
		t.Fatalf("GrandTotal details = %+v", grandTotal)
	}
}

func findFunctionReport(t *testing.T, functions []distance.FunctionReport, name string) distance.FunctionReport {
	t.Helper()

	for _, fn := range functions {
		if fn.Name == name {
			return fn
		}
	}

	t.Fatalf("function %s not found in %+v", name, functions)

	return distance.FunctionReport{}
}

func TestAnalyzePackageMetrics(t *testing.T) {
	report := analyzeFixture(t, nil)

	store := findPackage(t, report, "example.com/fixture/store")
	wantValue(t, store.Metrics, "distance", 0)
	if len(store.Metrics) != 1 || store.Metrics[0].Name != "distance" {
		t.Fatalf("store metrics = %v, want distance only", store.Metrics)
	}

	orders := findPackage(t, report, "example.com/fixture/orders")
	if orders.Vars != 0 || orders.Consts != 0 {
		t.Fatalf("orders counts = vars %d consts %d, want 0 and 0", orders.Vars, orders.Consts)
	}
	wantValue(t, orders.Metrics, "distance", 0)

	// An isolated package (Ca = Ce = 0) is defined as maximally stable:
	// instability 0, so distance = |0 + 0 − 1| = 1.
	isolated := findPackage(t, report, "example.com/fixture/isolated")
	wantValue(t, isolated.Metrics, "distance", 1)
}

func TestAnalyzeGeneratedFiles(t *testing.T) {
	report := analyzeFixture(t, nil)

	gen := findPackage(t, report, "example.com/fixture/gen")
	if len(gen.Types) != 0 {
		t.Fatalf("generated types analyzed by default: %v", gen.Types)
	}

	report = analyzeFixture(t, func(cfg *distance.Config) { cfg.IncludeGenerated = true })
	gen = findPackage(t, report, "example.com/fixture/gen")
	_ = findType(t, gen, "Machine")
}

func TestAnalyzeDeterminism(t *testing.T) {
	first := analyzeFixture(t, func(cfg *distance.Config) { cfg.Workers = 1 })

	second := analyzeFixture(t, func(cfg *distance.Config) { cfg.Workers = 8 })
	if !reflect.DeepEqual(first, second) {
		t.Fatal("reports differ across worker counts")
	}

	third := analyzeFixture(t, func(cfg *distance.Config) { cfg.Workers = 8 })
	if !reflect.DeepEqual(second, third) {
		t.Fatal("repeated runs differ")
	}
}

func TestAnalyzeInvalidConfig(t *testing.T) {
	ctx := context.Background()
	base := distance.Config{Directory: "../testdata/fixture"}

	bad := base

	bad.DependencyScope = "galaxy"
	if _, err := distance.Analyze(ctx, bad); err == nil {
		t.Fatal("invalid scope accepted")
	}
}

func TestAnalyzeCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := distance.Analyze(
		ctx,
		distance.Config{Directory: "../testdata/fixture"},
	); err == nil {
		t.Fatal("cancelled context accepted")
	}
}
