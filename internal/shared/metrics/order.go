package metrics

var reportedMetricOrder = []string{MetricAbstractness, MetricInstability, MetricDistance}

// ReportedMetricOrder is the public metric set this linter renders:
// abstractness, instability, then distance. Only distance is policy-gateable.
func ReportedMetricOrder() []string {
	return reportedMetricOrder
}
