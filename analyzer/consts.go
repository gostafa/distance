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

	keyDirectory       = "directory"
	keyDependencyScope = "dependency_scope"
	keyDependencyKebab = "dependency-scope"
	keyPatterns        = "patterns"
	keyRules           = "rules"
	keyBuildTags       = "build_tags"
	keyBuildTagsKebab  = "build-tags"
	keyWorkers         = "workers"
	keyTests           = "tests"
	keyGenerated       = "generated"
	keyContinueOnError = "continue_on_error"
	keyContinueKebab   = "continue-on-error"
	keyPattern         = "pattern"
	keyMax             = "max"
	errWrapValidate    = "analyzer validate: %w"
	errWrapNew         = "analyzer New: %w"
	errWrapAnalyze     = "analyzer analyze: %w"
	errWrapPolicy      = "distance policy: %w"
	errWrapAnalyzeRun  = "distance analyze: %w"
	errWrapSettings    = "analyzer Settings: %w"
	errWrapRule        = "analyzer rule: %w"
	errWrapApply       = "analyzer apply: %w"
)
