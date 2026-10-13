// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package embedding

func (w *Wrapper) Describe() string {
	w.Inc()             // promoted method: not a Wrapper method
	w.Base.Count += 1   // uses Wrapper's Base slot; Count stays Base's field
	if w.Count > zero { // promoted field: belongs to Base, not Wrapper
		return w.Name
	}
	return ""
}

func (b *Base) Inc() { b.Count++ }
