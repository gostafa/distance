// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package analyzer

const (
	// Name is the analyzer name registered with analysis and golangci-lint.
	Name = "distance"

	// Doc is the short analyzer documentation shown by go/analysis tools.
	Doc = `enforce Go package-distance policy thresholds

Reports policy violations when a package's distance from the main
sequence (|A + I − 1|) exceeds the first matching packages[] rule.
Configure rules inline in the golangci-lint settings block.`

	allPackagesPattern = "./..."

	floatBitSize   = 64
	floatPrecFixed = 2
	floatPrecAuto  = -1
	zero           = 0

	keyDirectory        = "directory"
	keyDependencyScope  = "dependency_scope"
	keyDependencyKebab  = "dependency-scope"
	keyPackages         = "packages"
	keyBuildTags        = "build_tags"
	keyBuildTagsKebab   = "build-tags"
	keyWorkers          = "workers"
	keyTests            = "tests"
	keyGenerated        = "generated"
	keyContinueOnError  = "continue_on_error"
	keyContinueKebab    = "continue-on-error"
	keyPattern          = "pattern"
	keyMaxDistance      = "max_distance"
	keyMaxDistanceKebab = "max-distance"
	errWrapValidate     = "analyzer validate: %w"
	errWrapNew          = "analyzer New: %w"
	errWrapAnalyze      = "analyzer analyze: %w"
	errWrapPolicy       = "distance policy: %w"
	errWrapAnalyzeRun   = "distance analyze: %w"
	errWrapSettings     = "analyzer Settings: %w"
	errWrapRule         = "analyzer package rule: %w"
	errWrapApply        = "analyzer apply: %w"
)
