// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Go.

package myers

// BitSet is a simple bit set implementation used by the Myers LCS algorithm.
type BitSet struct {
	bits []uint64
	size int
}

const elementSize = 64

func NewBitSet(size int) *BitSet {
	n := (size + elementSize - 1) / elementSize
	if n == 0 {
		n = 1
	}
	return &BitSet{bits: make([]uint64, n), size: size}
}

func (b *BitSet) Get(index int) bool {
	if index < 0 || index >= b.size {
		return false
	}
	return b.bits[index/elementSize]&(1<<uint(index%elementSize)) != 0
}

func (b *BitSet) Set(index int, value bool) {
	b.ensureCapacity(index)
	ei := index / elementSize
	mask := uint64(1) << uint(index%elementSize)
	if value {
		b.bits[ei] |= mask
	} else {
		b.bits[ei] &^= mask
	}
}

// SetRange sets bits from start (inclusive) to end (exclusive).
func (b *BitSet) SetRange(start, end int, value bool) {
	for i := start; i < end; i++ {
		b.Set(i, value)
	}
}

func (b *BitSet) ensureCapacity(index int) {
	if index >= b.size {
		b.size = index + 1
	}
	needed := (index / elementSize) + 1
	if needed > len(b.bits) {
		newBits := make([]uint64, needed)
		copy(newBits, b.bits)
		b.bits = newBits
	}
}
