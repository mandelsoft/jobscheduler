package utils

type entry[T any] struct {
	next *entry[T]
	elem T
}

// List is a simply linked list
// of elements.
type List[E any] struct {
	root *entry[E]
}

func (l *List[E]) IsEmpty() bool {
	return l.root == nil
}

func (l *List[E]) Last() E {
	var _nil E
	if l.root == nil {
		return _nil
	}
	p := &l.root
	for *p != nil {
		p = &(*p).next
	}
	return (*p).elem
}

func (l *List[E]) Last2() (E, bool) {
	var _nil E
	if l.root == nil {
		return _nil, false
	}
	p := &l.root
	for *p != nil {
		p = &(*p).next
	}
	return (*p).elem, true
}

func (l *List[E]) First() E {
	var _nil E
	if l.root == nil {
		return _nil
	}
	return l.root.elem
}

func (l *List[E]) First2() (E, bool) {
	var _nil E
	if l.root == nil {
		return _nil, false
	}
	return l.root.elem, true
}

func (l *List[E]) Remove(match func(e E) bool) bool {
	found := false
	p := &l.root
	for *p != nil {
		if match((*p).elem) {
			*p = (*p).next
			found = true
		} else {
			p = &(*p).next
		}
	}
	return found
}

func (l *List[E]) RemoveFirst() E {
	var _nil E
	if l.root == nil {
		return _nil
	}
	e := l.root.elem
	l.root = l.root.next
	return e
}

func (l *List[E]) RemoveFirst2() (E, bool) {
	var _nil E
	if l.root == nil {
		return _nil, false
	}
	e := l.root.elem
	l.root = l.root.next
	return e, true
}

func (l *List[E]) RemoveLast() E {
	var _nil E
	if l.root == nil {
		return _nil
	}
	p := &l.root
	for (*p).next != nil {
		p = &(*p).next
	}
	e := (*p).elem
	*p = nil
	return e
}

func (l *List[E]) RemoveLast2() (E, bool) {
	var _nil E
	if l.root == nil {
		return _nil, false
	}
	p := &l.root
	for (*p).next != nil {
		p = &(*p).next
	}
	e := (*p).elem
	*p = nil
	return e, true
}

func (l *List[E]) Append(e E) {
	p := &l.root
	for *p != nil {
		p = &(*p).next
	}
	*p = &entry[E]{elem: e}
}

func (l *List[E]) Insert(e E, before func(E) bool) {
	p := &l.root
	for *p != nil && !before((*p).elem) {
		p = &(*p).next
	}
	*p = &entry[E]{elem: e, next: *p}
}

func (l *List[E]) Prepend(e E) {
	l.root = &entry[E]{elem: e, next: l.root}
}
