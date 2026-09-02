package cli

import policydomain "github.com/gostafa/distance/internal/features/policy/domain"

// resolvePolicy builds a first-match policy from positional patterns and a
// shared maximum distance. Empty patterns default to "./...".
func resolvePolicy(patterns []string, maxDistance float64) (policydomain.Policy, string, error) {
	policy, err := policydomain.PolicyFromPatterns(patterns, maxDistance)
	if err != nil {
		return policydomain.Policy{}, "", err
	}

	return policy, "flag thresholds", nil
}
