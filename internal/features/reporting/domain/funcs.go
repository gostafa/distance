// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

// ParseFormat recognizes supported report output formats.
func ParseFormat(name string) (string, bool) {
	switch name {
	case "text", "json", "csv", "web":
		return name, true
	default:
		return "", false
	}
}
