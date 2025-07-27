package main

import "slices"

type LimitedSortedSlice[T any] struct {
	Compare func(left, right T) int
	Limit   int
	Arr     []T
	N       int
}

func (ls *LimitedSortedSlice[T]) Add(elem T) {
	ls.N++
	if len(ls.Arr) < ls.Limit {
		ls.Arr = append(ls.Arr, elem)
		slices.SortFunc(ls.Arr, ls.CompareDescending)
		return
	}

	if ls.Compare(elem, ls.Arr[len(ls.Arr)-1]) <= 0 {
		return
	}

	ls.Arr[len(ls.Arr)-1] = elem
	slices.SortFunc(ls.Arr, ls.CompareDescending)
}

func (ls *LimitedSortedSlice[T]) CompareDescending(a, b T) int {
	return -ls.Compare(a, b)
}
