package embedding

func (b *Base) Inc() { b.Count++ }

func (w *Wrapper) Describe() string {
	w.Inc()           // promoted method: not a Wrapper method
	w.Base.Count += 1 // uses Wrapper's Base slot; Count stays Base's field
	if w.Count > 0 {  // promoted field: belongs to Base, not Wrapper
		return w.Name
	}
	return ""
}
