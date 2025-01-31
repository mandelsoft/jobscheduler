package utils

type Entry[T any] struct {
	Next *Entry[T]
	Elem T
}
