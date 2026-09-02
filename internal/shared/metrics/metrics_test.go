package metrics

import (
	"math"
	"testing"
)

const epsilon = 1e-12

func assertApplicable(t *testing.T, r MetricResult, want float64) {
	t.Helper()

	if !r.Applicable {
		t.Fatalf("%s: not applicable (%s), want value %v", r.Name, r.Reason, want)
	}

	if math.Abs(r.Value-want) > epsilon {
		t.Fatalf("%s: value %v, want %v", r.Name, r.Value, want)
	}
}

func assertNotApplicable(t *testing.T, r MetricResult) {
	t.Helper()

	if r.Applicable {
		t.Fatalf("%s: applicable with value %v, want not applicable", r.Name, r.Value)
	}

	if r.Reason == "" {
		t.Fatalf("%s: not applicable without a reason", r.Name)
	}

	if r.Definition == "" {
		t.Fatalf("%s: missing definition", r.Name)
	}
}

func TestAbstractness(t *testing.T) {
	assertNotApplicable(t, Abstractness(0, 0))
	assertApplicable(t, Abstractness(1, 4), 0.25)
	assertApplicable(t, Abstractness(0, 3), 0)
}

func TestInstability(t *testing.T) {
	// Isolated package: defined as maximally stable, with a reason.
	isolated := Instability(0, 0)
	assertApplicable(t, isolated, 0)

	if isolated.Reason == "" {
		t.Fatal("isolated instability should carry the defined-as-0 reason")
	}

	assertApplicable(t, Instability(1, 0), 0)
	assertApplicable(t, Instability(0, 2), 1)
	assertApplicable(t, Instability(1, 3), 0.75)
}

func TestDistance(t *testing.T) {
	abstractness := Abstractness(1, 4) // 0.25
	instability := Instability(1, 3)   // 0.75
	assertApplicable(t, Distance(abstractness, instability), 0)

	assertNotApplicable(t, Distance(Abstractness(0, 0), instability))
	// Isolated packages have instability 0, so distance stays computable.
	assertApplicable(t, Distance(abstractness, Instability(0, 0)), 0.75)
	assertApplicable(t, Distance(Abstractness(1, 1), Instability(0, 2)), 1)

	assertNotApplicable(t, Distance(abstractness, MetricResult{
		Name: MetricInstability, Applicable: false, Reason: "no coupling data",
	}))
}

func TestReportedMetricOrder(t *testing.T) {
	got := ReportedMetricOrder()
	want := []string{MetricAbstractness, MetricInstability, MetricDistance}
	if len(got) != len(want) {
		t.Fatalf("ReportedMetricOrder() = %v, want %v", got, want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ReportedMetricOrder() = %v, want %v", got, want)
		}
	}
}
