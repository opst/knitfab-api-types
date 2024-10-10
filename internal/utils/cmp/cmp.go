package cmp

import "maps"

func SliceEqualUnordered[T interface{ Equal(T) bool }](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	// make a copy of b
	b = append([]T(nil), b...)

A:
	for _, x := range a {
		for i, y := range b {
			if x.Equal(y) {
				// remove y from b
				b = append(b[:i], b[i+1:]...)
				continue A
			}
		}
		return false
	}

	return len(b) == 0
}

func SliceEqEqUnordered[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	// make a copy of b
	b = append([]T(nil), b...)

A:
	for _, x := range a {
		for i, y := range b {
			if x == y {
				// remove y from b
				b = append(b[:i], b[i+1:]...)
				continue A
			}
		}
		return false
	}

	return len(b) == 0
}

func SliceEqual[T interface{ Equal(T) bool }](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	for i, x := range a {
		if !x.Equal(b[i]) {
			return false
		}
	}

	return true
}

func MapEqual[K comparable, V interface{ Equal(V) bool }](a, b map[K]V) bool {
	return MapEqualWith(a, b, V.Equal)
}

func MapEqualWith[K comparable, V any](a, b map[K]V, pred func(a, b V) bool) bool {
	if len(a) != len(b) {
		return false
	}

	// copy b
	b = maps.Clone(b)

	for k, va := range a {
		vb, ok := b[k]
		if !ok || !pred(va, vb) {
			return false
		}
		delete(b, k)
	}

	return len(b) == 0
}
