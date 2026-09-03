// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package cli

const (
	defaultWebReportName = "distance-report.html"

	exitUsage       = 2
	exitPolicy      = 3
	exitInterrupted = 130

	emptyString = ""
	zero        = 0
	one         = 1
	floatBits   = 64

	keyPath         = "path"
	keyError        = "error"
	keyPackages     = "packages"
	keyDuration     = "duration"
	keyWant         = "want"
	keyPolicyOrigin = "policy_origin"
	keyViolations   = "violations"
	keyFormat       = "format"

	flagNameRule          = "rule"
	helpTempPattern       = "distance-help-*.html"
	policySourceFlagRules = "flag rules"
	usageHeader           = "usage: distance [flags] [patterns...]\n\n"
	usageWebHint          = "\nFor an illustrated guide to the reported metrics:\n  distance --help --web\n"
	versionPrefix         = "distance "
	webFlagLong           = "--web"
	webFlagLongEq         = "--web="
	webFlagShort          = "-web"
	webFlagShortEq        = "-web="
	argTerminator         = "--"
	wantFormats           = "text, json, csv, or web"
	msgInvalidFormat      = "invalid format"
	msgConflictWeb        = "conflicting flags: -web implies -format=web"
	msgCPUFailed          = "cpu profiling failed"
	msgAnalysisFailed     = "analysis failed"
	msgAnalysisComplete   = "analysis complete"
	msgPolicyFailed       = "policy configuration failed"
	msgHeapFailed         = "memory profiling failed"
	msgWriteFailed        = "writing report failed"
	msgReportWritten      = "report written"
	msgPolicySucceeded    = "policy check succeeded"
	msgPolicyCheckFail    = "policy check failed"
	msgGuideFailed        = "writing the metrics guide failed"
	msgGuideWritten       = "metrics guide written"
	msgOpenReportFailed   = "opening the report in a browser failed"
	msgOpenGuideFailed    = "opening the metrics guide in a browser failed"
	errWrapWriteHelp      = "cli writeHelpDocs: %w"
	errWrapStderr         = "cli writeStderr: %w"
	errWrapPolicy         = "cli resolvePolicy: %w"
	errWrapPickOpen       = "cli pickOpen: %w"
	errFmtStdoutClose     = "stdout close: %w"
	errFmtStdoutWrite     = "stdout write: %w"
	errFmtStartCPU        = "profiling startCPU: %w"
	errFmtWriteHeap       = "profiling writeHeap: %w"
	errFmtCreateProfile   = "create profile file: %w"
	toolName              = "distance"
)
