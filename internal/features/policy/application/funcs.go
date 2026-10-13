// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"fmt"

	"github.com/gostafa/distance/internal/features/policy/domain"
)

// Check validates rules before applying the policy to an analysis report.
func Check(report *domain.Report, rules []domain.Rule) ([]domain.Violation, error) {
	err := domain.Validate(rules)
	if err != nil {
		return nil, fmt.Errorf("policy validation: %w", err)
	}

	return domain.Evaluate(report, rules), nil
}

// Evaluate applies a previously validated policy to the supplied report.
func Evaluate(report *domain.Report, rules []domain.Rule) []domain.Violation {
	return domain.Evaluate(report, rules)
}
