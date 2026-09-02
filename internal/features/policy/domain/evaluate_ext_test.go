package domain_test

import (
	"math"
	"strings"
	"testing"

	"github.com/gostafa/distance/internal/features/policy/domain"
	"github.com/gostafa/distance/internal/shared/metrics"
	"github.com/gostafa/distance/distance"
)

func sampleReport() distance.Report {
	return distance.Report{
		Packages: []distance.PackageReport{{
			Path:            "example.com/m/foo",
			Afferent:        2,
			Efferent:        20,
			ExportedFuncs:   20,
			UnexportedFuncs: 20,
			Vars:            20,
			Consts:          20,
			Functions: []distance.FunctionReport{{
				Name:  "Build",
				Lines: 100,
			}},
			Metrics: []metrics.MetricResult{
				{
					Name:       metrics.MetricDistance,
					Scope:      metrics.ScopePackage,
					Value:      0.9,
					Applicable: true,
				},
			},
			Types: []distance.TypeReport{{
				Name:    "Big",
				Fields:  20,
				Methods: 25,
				MethodDetails: []distance.FunctionReport{{
					Name:     "Do",
					Receiver: "Big",
					Lines:    90,
				}},
			}},
		}},
	}
}

func TestEvaluateFlagsMaxAndMinAndSkipsNotApplicable(t *testing.T) {
	t.Parallel()

	policy := domain.Policy{
		PackageMetrics: map[string]domain.Limit{
			metrics.MetricDistance: {Max: 0.8, HasMax: true},
		},
	}
	policy.Package.Efferent = domain.Limit{Max: 15, HasMax: true}
	policy.Package.Funcs.Count = domain.Limit{Max: 35, HasMax: true}
	policy.Package.Vars = domain.Limit{Max: 15, HasMax: true}
	policy.Package.Consts = domain.Limit{Max: 15, HasMax: true}
	policy.Type.Fields = domain.Limit{Max: 12, HasMax: true}
	policy.Funcs.Lines = domain.Limit{Max: 80, HasMax: true}

	got := domain.Evaluate(sampleReport(), policy)

	want := []struct {
		typ, fn, key string
		cmp          domain.Comparator
	}{
		{"", "", domain.KeyFuncs, domain.ComparatorMax},
		{"", "", domain.KeyVars, domain.ComparatorMax},
		{"", "", domain.KeyConsts, domain.ComparatorMax},
		{"", "", domain.KeyEfferent, domain.ComparatorMax},
		{"", "Build", domain.KeyFuncLines, domain.ComparatorMax},
		{"", "", metrics.MetricDistance, domain.ComparatorMax},
		{"Big", "", domain.KeyFields, domain.ComparatorMax},
		{"Big", "Do", domain.KeyFuncLines, domain.ComparatorMax},
	}

	if len(got) != len(want) {
		t.Fatalf("violations = %d, want %d\n%+v", len(got), len(want), got)
	}

	for i, w := range want {
		if got[i].Type != w.typ || got[i].Function != w.fn ||
			got[i].Key != w.key || got[i].Comparator != w.cmp {
			t.Errorf("violation[%d] = (%q %q %q %q), want (%q %q %q %q)",
				i,
				got[i].Type,
				got[i].Function,
				got[i].Key,
				got[i].Comparator,
				w.typ,
				w.fn,
				w.key,
				w.cmp,
			)
		}
	}
}

func TestEvaluateCleanReportHasNoViolations(t *testing.T) {
	t.Parallel()

	clean := distance.Report{
		Packages: []distance.PackageReport{{
			Path:          "example.com/m/tidy",
			ExportedFuncs: 3,
			Vars:          1,
			Consts:        1,
			Metrics: []metrics.MetricResult{{
				Name:       metrics.MetricDistance,
				Scope:      metrics.ScopePackage,
				Value:      0.1,
				Applicable: true,
			}},
			Types: []distance.TypeReport{{
				Name: "Small", Fields: 2, Methods: 2,
			}},
		}},
	}

	if got := domain.Evaluate(clean, domain.DefaultPolicy()); len(got) != 0 {
		t.Errorf("clean report produced violations: %+v", got)
	}
}

func TestEvaluateChecksBothBounds(t *testing.T) {
	t.Parallel()

	report := distance.Report{
		Packages: []distance.PackageReport{{
			Path: "p",
			Metrics: []metrics.MetricResult{
				{
					Name:       metrics.MetricDistance,
					Scope:      metrics.ScopePackage,
					Value:      0.9,
					Applicable: true,
				},
			},
		}},
	}
	policy := domain.Policy{Metrics: map[string]domain.Limit{
		metrics.MetricDistance: {Min: 0.1, HasMin: true, Max: 0.6, HasMax: true},
	}}

	got := domain.Evaluate(report, policy)
	if len(got) != 1 || got[0].Comparator != domain.ComparatorMax {
		t.Fatalf("want one max violation, got %+v", got)
	}
}

func TestEvaluateToleratesFloatingPointNoiseAtBoundary(t *testing.T) {
	t.Parallel()

	report := distance.Report{Packages: []distance.PackageReport{{
		Path: "p",
		Metrics: []metrics.MetricResult{
			{
				Name:       metrics.MetricDistance,
				Scope:      metrics.ScopePackage,
				Value:      math.Nextafter(0.5, 0),
				Applicable: true,
			},
		},
	}}}
	policy := domain.Policy{PackageMetrics: map[string]domain.Limit{
		metrics.MetricDistance: {Min: 0.5, HasMin: true},
	}}

	if got := domain.Evaluate(report, policy); len(got) != 0 {
		t.Fatalf("adjacent float below boundary produced violations: %+v", got)
	}

	report.Packages[0].Metrics[0].Value = 0.5 - 1e-9
	if got := domain.Evaluate(report, policy); len(got) != 1 {
		t.Fatalf("meaningful threshold crossing produced %d violations, want 1", len(got))
	}
}

func TestFormatViolations(t *testing.T) {
	t.Parallel()

	if s := domain.FormatViolations(nil); s != "" {
		t.Errorf("empty slice = %q, want empty", s)
	}

	out := domain.FormatViolations([]domain.Violation{
		{
			Package:    "example.com/m/foo",
			Key:        domain.KeyTypes,
			Value:      25,
			Comparator: domain.ComparatorMax,
			Threshold:  15,
		},
		{
			Package:    "example.com/m/foo",
			Function:   "Build",
			Key:        domain.KeyFuncLines,
			Value:      100,
			Comparator: domain.ComparatorMax,
			Threshold:  80,
		},
		{
			Package:    "example.com/m/foo",
			Key:        metrics.MetricDistance,
			Value:      0.9,
			Comparator: domain.ComparatorMax,
			Threshold:  0.5,
		},
	})

	for _, want := range []string{
		"policy: 3 violations",
		"example.com/m/foo (package): types 25 exceeds max 15",
		"example.com/m/foo.Build (func): funcs.lines 100 exceeds max 80",
		"example.com/m/foo (package): distance 0.90 exceeds max 0.50",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\ngot:\n%s", want, out)
		}
	}

	single := domain.FormatViolations(
		[]domain.Violation{
			{
				Package:    "p",
				Key:        domain.KeyExportedFuncs,
				Value:      5,
				Comparator: domain.ComparatorMax,
				Threshold:  3,
			},
		},
	)
	if !strings.HasPrefix(single, "policy: 1 violation\n") {
		t.Errorf("singular header wrong: %q", single)
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	if err := domain.Validate(domain.DefaultPolicy()); err != nil {
		t.Errorf("DefaultPolicy invalid: %v", err)
	}

	cases := map[string]domain.Policy{
		"unknown metric": {Metrics: map[string]domain.Limit{"nope": {Max: 1, HasMax: true}}},
		"min over max": {
			Metrics: map[string]domain.Limit{
				metrics.MetricDistance: {Min: 5, HasMin: true, Max: 2, HasMax: true},
			},
		},
		"hidden package metric": {
			PackageMetrics: map[string]domain.Limit{metrics.MetricAbstractness: {Max: 1, HasMax: true}},
		},
		"type metric rejected": {
			TypeMetrics: map[string]domain.Limit{metrics.MetricDistance: {Max: 1, HasMax: true}},
		},
	}
	for name, policy := range cases {
		if err := domain.Validate(policy); err == nil {
			t.Errorf("%s: want error, got nil", name)
		}
	}
}

func TestApplyOverride(t *testing.T) {
	t.Parallel()

	var policy domain.Policy

	if err := domain.ApplyOverride(&policy, domain.KeyTypes, domain.ComparatorMax, 10); err != nil {
		t.Fatal(err)
	}

	if err := domain.ApplyOverride(
		&policy,
		metrics.MetricDistance,
		domain.ComparatorMax,
		0.8,
	); err != nil {
		t.Fatal(err)
	}

	if err := domain.ApplyOverride(
		&policy,
		"package."+metrics.MetricDistance,
		domain.ComparatorMax,
		0.5,
	); err != nil {
		t.Fatal(err)
	}

	if err := domain.ApplyOverride(&policy, "bogus", domain.ComparatorMax, 1); err == nil {
		t.Error("unknown key: want error, got nil")
	}

	if err := domain.ApplyOverride(&policy, "type."+metrics.MetricDistance, domain.ComparatorMax, 1); err == nil {
		t.Error("type.distance: want error, got nil")
	}

	if err := domain.ApplyOverride(&policy, "package."+metrics.MetricAbstractness, domain.ComparatorMax, 1); err == nil {
		t.Error("package.abstractness: want error, got nil")
	}

	if !policy.Package.Types.HasMax || policy.Package.Types.Max != 10 {
		t.Errorf("types override not applied: %+v", policy.Package.Types)
	}
	if err := domain.ApplyOverride(&policy, domain.KeyFuncs, domain.ComparatorMax, 5); err != nil {
		t.Fatal(err)
	}
	if err := domain.ApplyOverride(&policy, domain.KeyVars, domain.ComparatorMax, 6); err != nil {
		t.Fatal(err)
	}
	if err := domain.ApplyOverride(&policy, domain.KeyConsts, domain.ComparatorMax, 7); err != nil {
		t.Fatal(err)
	}
	if !policy.Package.Vars.HasMax || policy.Package.Vars.Max != 6 {
		t.Errorf("vars override not applied: %+v", policy.Package.Vars)
	}
	if !policy.Package.Funcs.Count.HasMax || policy.Package.Funcs.Count.Max != 5 {
		t.Errorf("funcs override not applied: %+v", policy.Package.Funcs)
	}
	if !policy.Package.Consts.HasMax || policy.Package.Consts.Max != 7 {
		t.Errorf("consts override not applied: %+v", policy.Package.Consts)
	}
	if err := domain.ApplyOverride(&policy, domain.KeyFuncLines, domain.ComparatorMax, 80); err != nil {
		t.Fatal(err)
	}
	if !policy.Funcs.Lines.HasMax || policy.Funcs.Lines.Max != 80 {
		t.Errorf("funcs.lines override not applied: %+v", policy.Funcs)
	}

	if l := policy.Metrics[metrics.MetricDistance]; !l.HasMax || l.Max != 0.8 {
		t.Errorf("distance override not applied: %+v", l)
	}

	if l := policy.PackageMetrics[metrics.MetricDistance]; !l.HasMax || l.Max != 0.5 {
		t.Errorf("package.distance override not applied: %+v", l)
	}
}

func TestMetricNamesSorted(t *testing.T) {
	t.Parallel()

	policy := domain.Policy{Metrics: map[string]domain.Limit{
		metrics.MetricDistance: {Min: 0.3, HasMin: true},
	}, PackageMetrics: map[string]domain.Limit{
		metrics.MetricDistance: {Max: 0.4, HasMax: true},
	}}

	got := domain.MetricNames(policy)
	if len(got) != 1 || got[0] != metrics.MetricDistance {
		t.Fatalf("names = %v, want [%s]", got, metrics.MetricDistance)
	}
}
