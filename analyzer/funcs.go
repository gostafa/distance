// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package analyzer

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"go/token"
	"math"
	"reflect"
	"strconv"

	"github.com/gostafa/distance/distance"
	policydomain "github.com/gostafa/distance/internal/features/policy/domain"
	"golang.org/x/tools/go/analysis"
)

// New builds an analysis.Analyzer from the given Settings.
func New(settings *Settings) (*analysis.Analyzer, error) {
	validated := settings.withDefaults()

	validateErr := validated.validate()
	if validateErr != nil {
		return nil, fmt.Errorf(errWrapNew, validateErr)
	}

	return bindAnalyzer(validated), nil
}

func (fn analyzeFunc) Analyze(ctx context.Context, cfg *distance.Config) (distance.Report, error) {
	report, analyzeErr := fn(ctx, cfg)
	if analyzeErr != nil {
		return distance.Report{}, fmt.Errorf(errWrapAnalyze, analyzeErr)
	}

	return report, nil
}

func applyBool(raw json.RawMessage, dest *bool) error {
	if raw == nil {
		return nil
	}

	unmarshalErr := json.Unmarshal(raw, dest)
	if unmarshalErr != nil {
		return fmt.Errorf(errWrapSettings, unmarshalErr)
	}

	return nil
}

func applyInt(raw json.RawMessage, dest *int) error {
	if raw == nil {
		return nil
	}

	unmarshalErr := json.Unmarshal(raw, dest)
	if unmarshalErr != nil {
		return fmt.Errorf(errWrapSettings, unmarshalErr)
	}

	return nil
}

func applyString(raw json.RawMessage, dest *string) error {
	if raw == nil {
		return nil
	}

	unmarshalErr := json.Unmarshal(raw, dest)
	if unmarshalErr != nil {
		return fmt.Errorf(errWrapSettings, unmarshalErr)
	}

	return nil
}

func applyStrings(raw json.RawMessage, dest *[]string) error {
	if raw == nil {
		return nil
	}

	unmarshalErr := json.Unmarshal(raw, dest)
	if unmarshalErr != nil {
		return fmt.Errorf(errWrapSettings, unmarshalErr)
	}

	return nil
}

func bindAnalyzer(settings *Settings) *analysis.Analyzer {
	built := &analysis.Analyzer{
		Name:       Name,
		Doc:        Doc,
		ResultType: reflect.TypeFor[*runResult](),
	}

	bindRun(built, newRunner(settings))

	return built
}

func computeViolations(cfg *Settings, src reportAnalyzer) (pkgViolations, error) {
	policy, policyErr := cfg.policy()
	if policyErr != nil {
		return nil, fmt.Errorf(errWrapPolicy, policyErr)
	}

	report, analyzeErr := src.Analyze(context.Background(), cfg.toConfig())
	if analyzeErr != nil {
		return nil, fmt.Errorf(errWrapAnalyzeRun, analyzeErr)
	}

	return groupByPackage(policydomain.Evaluate(&report, policy)), nil
}

func decodePackageRule(raw json.RawMessage) (policydomain.PackageRule, error) {
	fields, unmarshalErr := unmarshalObject(raw)
	if unmarshalErr != nil {
		return policydomain.PackageRule{}, fmt.Errorf(errWrapRule, unmarshalErr)
	}

	rule, fromErr := packageRuleFromFields(fields)
	if fromErr != nil {
		return policydomain.PackageRule{}, fmt.Errorf(errWrapApply, fromErr)
	}

	return rule, nil
}

func unmarshalObject(raw json.RawMessage) (map[string]json.RawMessage, error) {
	fields := map[string]json.RawMessage{}

	unmarshalErr := json.Unmarshal(raw, &fields)
	if unmarshalErr != nil {
		return nil, fmt.Errorf(errWrapSettings, unmarshalErr)
	}

	return fields, nil
}

func packageRuleFromFields(fields map[string]json.RawMessage) (policydomain.PackageRule, error) {
	unknownErr := rejectUnknown(fields, ruleKeys())
	if unknownErr != nil {
		return policydomain.PackageRule{}, fmt.Errorf(errWrapApply, unknownErr)
	}

	rule, fromErr := packageRuleFrom(fields)
	if fromErr != nil {
		return policydomain.PackageRule{}, fmt.Errorf(errWrapApply, fromErr)
	}

	return rule, nil
}

func decodePackages(raw json.RawMessage) ([]policydomain.PackageRule, error) {
	if raw == nil {
		return nil, nil
	}

	var items []json.RawMessage

	unmarshalErr := json.Unmarshal(raw, &items)
	if unmarshalErr != nil {
		return nil, fmt.Errorf(errWrapSettings, unmarshalErr)
	}

	rules, listErr := decodeRuleList(items)
	if listErr != nil {
		return nil, fmt.Errorf(errWrapApply, listErr)
	}

	return rules, nil
}

func decodeRuleList(items []json.RawMessage) ([]policydomain.PackageRule, error) {
	rules := make([]policydomain.PackageRule, zero, len(items))

	for i := range items {
		rule, decodeErr := decodePackageRule(items[i])
		if decodeErr != nil {
			return nil, fmt.Errorf(errWrapApply, decodeErr)
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

func firstRaw(raw map[string]json.RawMessage, keys ...string) json.RawMessage {
	for i := range keys {
		value, ok := raw[keys[i]]

		if ok {
			return value
		}
	}

	return nil
}

func formatNumber(value float64) string {
	if value == math.Trunc(value) && !math.IsInf(value, zero) {
		return strconv.FormatFloat(value, 'f', floatPrecAuto, floatBitSize)
	}

	return strconv.FormatFloat(value, 'f', floatPrecFixed, floatBitSize)
}

func formatViolation(violation *policydomain.Violation) string {
	return fmt.Sprintf(
		"%s (package): %s %s exceeds max %s",
		violation.Package,
		violation.Key,
		formatNumber(violation.Value),
		formatNumber(violation.Threshold),
	)
}

func groupByPackage(violations []policydomain.Violation) map[string][]policydomain.Violation {
	byPkg := make(map[string][]policydomain.Violation, len(violations))

	for i := range violations {
		violation := &violations[i]

		byPkg[violation.Package] = append(byPkg[violation.Package], *violation)
	}

	return byPkg
}

func newRunner(settings *Settings) *runner {
	return &runner{settings: settings, analyzer: analyzeFunc(distance.Analyze)}
}

func packagePos(pass *analysis.Pass) token.Pos {
	for i := range pass.Files {
		if pass.Files[i] != nil {
			return pass.Files[i].Package
		}
	}

	return token.NoPos
}

func packageRuleFrom(fields map[string]json.RawMessage) (policydomain.PackageRule, error) {
	var rule policydomain.PackageRule

	patternErr := applyString(firstRaw(fields, keyPattern), &rule.Pattern)
	if patternErr != nil {
		return policydomain.PackageRule{}, fmt.Errorf(errWrapApply, patternErr)
	}

	numberErr := applyNumber(
		firstRaw(fields, keyMaxDistance, keyMaxDistanceKebab),
		&rule.MaxDistance,
	)
	if numberErr != nil {
		return policydomain.PackageRule{}, fmt.Errorf(errWrapApply, numberErr)
	}

	return rule, nil
}

func applyNumber(raw json.RawMessage, dest *float64) error {
	if raw == nil {
		return nil
	}

	unmarshalErr := json.Unmarshal(raw, dest)
	if unmarshalErr != nil {
		return fmt.Errorf(errWrapRule, unmarshalErr)
	}

	return nil
}

func reportViolations(pass *analysis.Pass, violations []policydomain.Violation) {
	for i := range violations {
		violation := &violations[i]

		pass.Report(analysis.Diagnostic{
			Pos:      packagePos(pass),
			Category: Name,
			Message:  formatViolation(violation),
		})
	}
}

func (runner *runner) load() {
	runner.byPkg, runner.err = computeViolations(runner.settings, runner.analyzer)
}

func (runner *runner) run(pass *analysis.Pass) (*runResult, error) {
	runner.once.Do(runner.load)

	if runner.err != nil {
		return nil, runner.err
	}

	reportViolations(pass, runner.byPkg[pass.Pkg.Path()])

	return &runResult{}, nil
}

func (err scopeError) Error() string {
	return fmt.Sprintf(
		"invalid dependency-scope %q (want project, module, or all)",
		err.value,
	)
}

func (settings *Settings) policy() (policydomain.Policy, error) {
	policy := policydomain.Policy{Packages: settings.Packages}

	validateErr := policydomain.Validate(policy)
	if validateErr != nil {
		return policydomain.Policy{}, fmt.Errorf("analyzer policy: %w", validateErr)
	}

	return policy, nil
}

func (settings *Settings) toConfig() *distance.Config {
	return &distance.Config{
		Directory:        settings.Directory,
		Patterns:         policydomain.LoadPatterns(settings.Packages),
		IncludeTests:     settings.Tests,
		IncludeGenerated: settings.Generated,
		BuildTags:        append([]string(nil), settings.BuildTags...),
		Workers:          settings.Workers,
		DependencyScope:  distance.DependencyScope(settings.DependencyScope),
		ContinueOnError:  settings.ContinueOnError,
	}
}

func (err unknownFieldError) Error() string {
	return "unknown settings key " + err.key
}

// UnmarshalJSON decodes analyzer settings from kebab-case or snake_case JSON.
func (settings *Settings) UnmarshalJSON(data []byte) error {
	raw, unmarshalErr := unmarshalObject(data)
	if unmarshalErr != nil {
		return fmt.Errorf(errWrapSettings, unmarshalErr)
	}

	applyErr := settings.applyDecoded(raw)
	if applyErr != nil {
		return fmt.Errorf(errWrapApply, applyErr)
	}

	return nil
}

func (settings *Settings) applyDecoded(raw map[string]json.RawMessage) error {
	unknownErr := rejectUnknown(raw, settingsKeys())
	if unknownErr != nil {
		return fmt.Errorf(errWrapApply, unknownErr)
	}

	applyErr := settings.applyRaw(raw)
	if applyErr != nil {
		return fmt.Errorf(errWrapApply, applyErr)
	}

	return nil
}

func rejectUnknown(raw map[string]json.RawMessage, known map[string]bool) error {
	for key := range raw {
		if !known[key] {
			return unknownFieldError{key: key}
		}
	}

	return nil
}

func settingsKeys() map[string]bool {
	return map[string]bool{
		keyDirectory:       true,
		keyDependencyScope: true,
		keyDependencyKebab: true,
		keyPackages:        true,
		keyBuildTags:       true,
		keyBuildTagsKebab:  true,
		keyWorkers:         true,
		keyTests:           true,
		keyGenerated:       true,
		keyContinueOnError: true,
		keyContinueKebab:   true,
	}
}

func ruleKeys() map[string]bool {
	return map[string]bool{
		keyPattern:          true,
		keyMaxDistance:      true,
		keyMaxDistanceKebab: true,
	}
}

func (settings *Settings) applyRaw(raw map[string]json.RawMessage) error {
	scalarErr := settings.applyScalars(raw)
	if scalarErr != nil {
		return fmt.Errorf(errWrapApply, scalarErr)
	}

	listErr := settings.applyLists(raw)
	if listErr != nil {
		return fmt.Errorf(errWrapApply, listErr)
	}

	return nil
}

func (settings *Settings) applyLists(raw map[string]json.RawMessage) error {
	tagsErr := applyStrings(firstRaw(raw, keyBuildTags, keyBuildTagsKebab), &settings.BuildTags)
	if tagsErr != nil {
		return fmt.Errorf(errWrapApply, tagsErr)
	}

	rules, packagesErr := decodePackages(firstRaw(raw, keyPackages))
	if packagesErr != nil {
		return fmt.Errorf(errWrapApply, packagesErr)
	}

	settings.Packages = rules

	return nil
}

func (settings *Settings) applyScalars(raw map[string]json.RawMessage) error {
	dirErr := applyString(firstRaw(raw, keyDirectory), &settings.Directory)
	if dirErr != nil {
		return fmt.Errorf(errWrapApply, dirErr)
	}

	scopeErr := applyString(
		firstRaw(raw, keyDependencyScope, keyDependencyKebab),
		&settings.DependencyScope,
	)
	if scopeErr != nil {
		return fmt.Errorf(errWrapApply, scopeErr)
	}

	flagsErr := settings.applyFlags(raw)
	if flagsErr != nil {
		return fmt.Errorf(errWrapApply, flagsErr)
	}

	return nil
}

func (settings *Settings) applyFlags(raw map[string]json.RawMessage) error {
	workersErr := applyInt(firstRaw(raw, keyWorkers), &settings.Workers)
	if workersErr != nil {
		return fmt.Errorf(errWrapApply, workersErr)
	}

	flagsErr := settings.applyBoolFlags(raw)
	if flagsErr != nil {
		return fmt.Errorf(errWrapApply, flagsErr)
	}

	return nil
}

func (settings *Settings) applyBoolFlags(raw map[string]json.RawMessage) error {
	testsErr := applyBool(firstRaw(raw, keyTests), &settings.Tests)
	if testsErr != nil {
		return fmt.Errorf(errWrapApply, testsErr)
	}

	generatedErr := applyBool(firstRaw(raw, keyGenerated), &settings.Generated)
	if generatedErr != nil {
		return fmt.Errorf(errWrapApply, generatedErr)
	}

	continueErr := applyBool(
		firstRaw(raw, keyContinueOnError, keyContinueKebab),
		&settings.ContinueOnError,
	)
	if continueErr != nil {
		return fmt.Errorf(errWrapApply, continueErr)
	}

	return nil
}

func (settings *Settings) validate() error {
	scopeErr := validateDependencyScope(settings.DependencyScope)
	if scopeErr != nil {
		return fmt.Errorf(errWrapValidate, scopeErr)
	}

	policy, policyErr := settings.policy()
	if policyErr != nil {
		return fmt.Errorf(errWrapValidate, policyErr)
	}

	settings.Packages = policy.Packages

	return nil
}

func (settings *Settings) withDefaults() *Settings {
	if len(settings.Packages) == zero {
		settings.Packages = []policydomain.PackageRule{{
			Pattern:     allPackagesPattern,
			MaxDistance: policydomain.DefaultMaxDistance,
		}}
	}

	settings.DependencyScope = cmp.Or(
		settings.DependencyScope,
		string(distance.DependencyScopeModule),
	)

	return settings
}

func validateDependencyScope(value string) error {
	switch distance.DependencyScope(value) {
	case distance.DependencyScopeProject,
		distance.DependencyScopeModule,
		distance.DependencyScopeAll:
		return nil
	default:
		return scopeError{value: value}
	}
}
