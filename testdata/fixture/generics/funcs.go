package generics

func (p *Pair[T]) SetFirst(v T) { p.First = v }

func (p *Pair[T]) SetSecond(v T) { p.Second = v }

func (p Pair[T]) Swapped(other Pair[T]) Pair[T] {
	return Pair[T]{First: other.Second, Second: other.First}
}
