package common

// List is a generic container that holds a sequence of items of any type.
// It provides methods to add items and retrieve the underlying slice.
type List[T any] struct {
	items []T
}

func (l *List[T]) Add(item T) {
	l.items = append(l.items, item)
}

func (l *List[T]) Items() []T {
	return l.items
}
