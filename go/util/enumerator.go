// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Go.

package util

// Enumerator assigns unique integer IDs to strings.
// Equal strings receive the same ID. IDs start from 1; 0 is reserved for empty.
type Enumerator struct {
	numbers    map[string]int
	nextNumber int
}

func NewEnumerator(expectedCapacity int) *Enumerator {
	return &Enumerator{
		numbers:    make(map[string]int, expectedCapacity),
		nextNumber: 1,
	}
}

// Enumerate assigns integer IDs to each string in the slice.
func (e *Enumerator) Enumerate(objects []string) []int {
	result := make([]int, len(objects))
	for i, obj := range objects {
		result[i] = e.enumerate(obj)
	}
	return result
}

func (e *Enumerator) enumerate(obj string) int {
	if n, ok := e.numbers[obj]; ok {
		return n
	}
	n := e.nextNumber
	e.nextNumber++
	e.numbers[obj] = n
	return n
}
