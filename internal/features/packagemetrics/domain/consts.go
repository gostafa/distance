// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

const (
	// ScopeProject counts only imports of other analyzed packages.
	ScopeProject = "project"
	// ScopeModule counts imports of packages in the main module. Without
	// module information it degrades to ScopeProject.
	ScopeModule = "module"
	// ScopeAll counts every import.
	ScopeAll = "all"

	zero = 0
)
