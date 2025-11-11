package common

type List[T any] struct {
	items []T
}

func (l *List[T]) Add(item T) {
	l.items = append(l.items, item)
}

func (l *List[T]) Items() []T {
	return l.items
}
