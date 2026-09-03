// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package analyzer

const (
	// Name is the analyzer name registered with analysis and golangci-lint.
	Name = "distance"

	// Doc is the short analyzer documentation shown by go/analysis tools.
	Doc = `enforce Go package-distance policy thresholds

Reports policy violations when a package's distance from the main
sequence (|A + I − 1|) exceeds the most specific matching rules[] max.
Rules match package import paths using glob patterns (* = one segment,
** = zero or more). Load patterns are independent of policy rules.`

	defaultPackagePattern = "./..."

	floatBitSize   = 64
	floatPrecFixed = 2
	floatPrecAuto  = -1
	zero           = 0

	errFmtUnmarshal   = "UnmarshalSettings: %w"
	errFmtValidate    = "validate: %w"
	errFmtRemap       = "remapKebabKeys: %w"
	errWrapNew        = "analyzer New: %w"
	errWrapPolicy     = "distance policy: %w"
	errWrapAnalyzeRun = "distance analyze: %w"
)
