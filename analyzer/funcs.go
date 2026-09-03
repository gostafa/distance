// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package analyzer

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"go/token"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/gostafa/distance/distance"
	"github.com/gostafa/distance/distance/wire"
	policydomain "github.com/gostafa/distance/internal/features/policy/domain"
	"golang.org/x/tools/go/analysis"
)

// New returns a go/analysis Analyzer that loads the module once, evaluates the
// distance policy, and emits diagnostics for the package under analysis.
func New(settings *Settings) (*analysis.Analyzer, error) {
	resolved := settingsWithDefaults(settings)

	err := validateSettings(&resolved)
	if err != nil {
		return nil, fmt.Errorf(errWrapNew, err)
	}

	active := newRunner(&resolved)

	return &analysis.Analyzer{
		Name:       Name,
		Doc:        Doc,
		ResultType: reflect.TypeFor[*runResult](),
		Run:        func(pass *analysis.Pass) (any, error) { return runRunner(active, pass) },
	}, nil
}

// UnmarshalSettings accepts snake_case tags and remaps kebab-case keys from
// golangci-lint settings so DisallowUnknownFields still applies.
func UnmarshalSettings(data []byte, settings *Settings) error {
	err := decodeUnmarshaledSettings(settings, data)
	if err != nil {
		return fmt.Errorf("decode settings: %w", err)
	}

	return nil
}

func decodeUnmarshaledSettings(settings *Settings, data []byte) error {
	remapped, err := remapKebabKeys(data)
	if err != nil {
		return fmt.Errorf(errFmtUnmarshal, err)
	}

	err = decodeSettings(settings, remapped)
	if err != nil {
		return fmt.Errorf(errFmtUnmarshal, err)
	}

	return nil
}

func decodeSettings(settings *Settings, data []byte) error {
	type settingsAlias Settings

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var alias settingsAlias

	err := decoder.Decode(&alias)
	if err != nil {
		return fmt.Errorf(errFmtUnmarshal, err)
	}

	*settings = Settings(alias)

	return nil
}

func settingsRules(settings *Settings) ([]policydomain.Rule, error) {
	if len(settings.Rules) == zero {
		return policydomain.DefaultRules(), nil
	}

	parsed, err := parseSettingsRules(settings)
	if err != nil {
		return nil, fmt.Errorf("rules: %w", err)
	}

	return parsed, nil
}

func parseSettingsRules(settings *Settings) ([]policydomain.Rule, error) {
	rules := make([]policydomain.Rule, zero, len(settings.Rules))

	for index := range settings.Rules {
		if settings.Rules[index].Max == nil {
			return nil, fmt.Errorf("parseRules: rules[%d]: %w", index, errRuleMaxRequired)
		}

		rules = append(rules, policydomain.Rule{
			Pattern: settings.Rules[index].Pattern,
			Max:     *settings.Rules[index].Max,
		})
	}

	err := policydomain.Validate(rules)
	if err != nil {
		return nil, fmt.Errorf("parseRules: %w", err)
	}

	return rules, nil
}

func settingsToConfig(settings *Settings) distance.Config {
	return distance.Config{
		Directory:        settings.Directory,
		Patterns:         append([]string(nil), settings.Patterns...),
		IncludeTests:     settings.Tests,
		IncludeGenerated: settings.Generated,
		BuildTags:        append([]string(nil), settings.BuildTags...),
		Workers:          settings.Workers,
		DependencyScope:  settings.DependencyScope,
		ContinueOnError:  settings.ContinueOnError,
	}
}

func validateSettings(settings *Settings) error {
	err := validateDependencyScope(settings.DependencyScope)
	if err != nil {
		return fmt.Errorf(errFmtValidate, err)
	}

	rules, err := settingsRules(settings)
	if err != nil {
		return fmt.Errorf(errFmtValidate, err)
	}

	err = policydomain.Validate(rules)
	if err != nil {
		return fmt.Errorf(errFmtValidate, err)
	}

	return nil
}

func settingsWithDefaults(settings *Settings) Settings {
	out := *settings

	if len(out.Patterns) == zero {
		out.Patterns = []string{defaultPackagePattern}
	}

	out.DependencyScope = cmp.Or(out.DependencyScope, distance.DependencyScopeModule)

	return out
}

func computeViolations(settings *Settings, analyzer reportAnalyzer) (pkgViolations, error) {
	rules, err := settingsRules(settings)
	if err != nil {
		return nil, fmt.Errorf(errWrapPolicy, err)
	}

	cfg := settingsToConfig(settings)

	report, err := analyzer(context.Background(), &cfg)
	if err != nil {
		return nil, fmt.Errorf(errWrapAnalyzeRun, err)
	}

	return groupByPackage(policydomain.Evaluate(&report, rules)), nil
}

func formatNumber(value float64) string {
	if value == math.Trunc(value) && !math.IsInf(value, zero) {
		return strconv.FormatFloat(value, 'f', floatPrecAuto, floatBitSize)
	}

	return strconv.FormatFloat(value, 'f', floatPrecFixed, floatBitSize)
}

func formatViolation(violation *policydomain.Violation) string {
	return fmt.Sprintf(
		"%s (package): %s %s exceeds max %s (rule %s)",
		violation.Package,
		distance.MetricDistance,
		formatNumber(violation.Value),
		formatNumber(violation.Threshold),
		violation.Rule,
	)
}

func groupByPackage(violations []policydomain.Violation) map[string][]policydomain.Violation {
	byPkg := make(map[string][]policydomain.Violation, len(violations))

	for index := range violations {
		violation := &violations[index]

		byPkg[violation.Package] = append(byPkg[violation.Package], *violation)
	}

	return byPkg
}

func newRunner(settings *Settings) *runner {
	return &runner{settings: *settings, analyzer: wire.AnalyzeWithDefault}
}

func packagePos(pass *analysis.Pass) token.Pos {
	for index := range pass.Files {
		if pass.Files[index] != nil {
			return pass.Files[index].Package
		}
	}

	return token.NoPos
}

func remapKebabKeys(data []byte) ([]byte, error) {
	raw, err := unmarshalSettingsMap(data)
	if err != nil {
		return nil, fmt.Errorf(errFmtRemap, err)
	}

	encoded, err := marshalRemappedKeys(raw)
	if err != nil {
		return nil, fmt.Errorf(errFmtRemap, err)
	}

	return encoded, nil
}

func unmarshalSettingsMap(data []byte) (map[string]json.RawMessage, error) {
	var raw map[string]json.RawMessage

	err := json.Unmarshal(data, &raw)
	if err != nil {
		return nil, fmt.Errorf(errFmtRemap, err)
	}

	return raw, nil
}

func marshalRemappedKeys(raw map[string]json.RawMessage) ([]byte, error) {
	out := make(map[string]json.RawMessage, len(raw))

	for key := range raw {
		out[strings.ReplaceAll(key, "-", "_")] = raw[key]
	}

	encoded, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf(errFmtRemap, err)
	}

	return encoded, nil
}

func reportViolations(pass *analysis.Pass, violations []policydomain.Violation) {
	for index := range violations {
		violation := &violations[index]

		pass.Report(analysis.Diagnostic{
			Pos:      packagePos(pass),
			Category: Name,
			Message:  formatViolation(violation),
		})
	}
}

func loadRunner(state *runner) {
	state.byPkg, state.err = computeViolations(&state.settings, state.analyzer)
}

func runRunner(state *runner, pass *analysis.Pass) (*runResult, error) {
	state.once.Do(func() { loadRunner(state) })

	if state.err != nil {
		return nil, fmt.Errorf("run: %w", state.err)
	}

	reportViolations(pass, state.byPkg[pass.Pkg.Path()])

	return &runResult{}, nil
}

func (err scopeError) Error() string {
	return fmt.Sprintf(
		"invalid dependency-scope %q (want project, module, or all)",
		err.value,
	)
}

func validateDependencyScope(value string) error {
	switch value {
	case distance.DependencyScopeProject,
		distance.DependencyScopeModule,
		distance.DependencyScopeAll:
		return nil
	default:
		return scopeError{value: value}
	}
}
