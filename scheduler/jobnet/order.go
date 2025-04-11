package jobnet

import (
	"slices"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/goutils/sliceutils"
)

func order[T comparable](elems map[T]set.Set[T]) ([]T, [][]T) {
	var order []T
	var cycles [][]T
	for e := range root(elems) {
		o, c := addDeps(e, nil, order, elems)
		order = o
		cycles = append(cycles, c...)
	}
	return order, cycles
}

func root[T comparable](elems map[T]set.Set[T]) set.Set[T] {
	set := set.KeySet(elems)

	for _, deps := range elems {
		set.DeleteAll(deps)
	}
	return set
}

func addDeps[T comparable](e T, stack []T, ordered []T, elems map[T]set.Set[T]) ([]T, [][]T) {
	var cycles [][]T

	if cycle := general.Cycle(e, stack...); cycle != nil {
		return nil, [][]T{cycle}
	}
	if slices.Contains(ordered, e) {
		return ordered, nil
	}
	stack = sliceutils.CopyAppend(stack, e)
	for d := range elems[e] {
		var nested [][]T
		ordered, nested = addDeps[T](d, stack, ordered, elems)
		cycles = append(cycles, nested...)
	}
	ordered = append(ordered, e)
	return ordered, cycles
}
