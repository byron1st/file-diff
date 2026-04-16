// Copyright 2000-2025 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.
// Ported from intellij-community to Go.

package myers

import (
	"errors"
	"math"
)

// ErrTooBig is returned when the diff is too large to compute.
var ErrTooBig = errors.New("diff too big")

// MyersLCS finds the longest common subsequence of two integer arrays
// using the Myers O(ND) algorithm.
//
// Based on E.W. Myers / An O(ND) Difference Algorithm and Its Variations / 1986
type MyersLCS struct {
	first  []int
	second []int

	start1 int
	count1 int
	start2 int
	count2 int

	changes1 *BitSet
	changes2 *BitSet

	vForward  []int
	vBackward []int
}

func NewMyersLCS(first, second []int) *MyersLCS {
	m := &MyersLCS{
		first:    first,
		second:   second,
		start1:   0,
		count1:   len(first),
		start2:   0,
		count2:   len(second),
		changes1: NewBitSet(len(first)),
		changes2: NewBitSet(len(second)),
	}
	m.changes1.SetRange(0, len(first), true)
	m.changes2.SetRange(0, len(second), true)

	total := m.count1 + m.count2
	m.vForward = make([]int, total+1)
	m.vBackward = make([]int, total+1)
	return m
}

// Execute runs the full Myers algorithm without any threshold.
func (m *MyersLCS) Execute() {
	threshold := m.count1 + m.count2
	_ = m.execute(threshold, false)
}

// ExecuteWithThreshold runs the Myers algorithm with a threshold.
// Returns ErrTooBig if the threshold is exceeded.
func (m *MyersLCS) ExecuteWithThreshold() error {
	threshold := max(20000+10*int(math.Sqrt(float64(m.count1+m.count2))), deltaThresholdSize)
	return m.execute(threshold, true)
}

func (m *MyersLCS) execute(threshold int, throwErr bool) error {
	if m.count1 == 0 || m.count2 == 0 {
		return nil
	}
	threshold = min(threshold, m.count1+m.count2)
	return m.run(0, m.count1, 0, m.count2, threshold, throwErr)
}

func (m *MyersLCS) run(oldStart, oldEnd, newStart, newEnd, diffEstimate int, throwErr bool) error {
	if oldStart >= oldEnd || newStart >= newEnd {
		return nil
	}

	oldLength := oldEnd - oldStart
	newLength := newEnd - newStart
	m.vForward[newLength+1] = 0
	m.vBackward[newLength+1] = 0
	halfD := (diffEstimate + 1) / 2

	xx, kk, td := -1, -1, -1

outer:
	for d := 0; d <= halfD; d++ {
		L := newLength + max(-d, -newLength+((d^newLength)&1))
		R := newLength + min(d, oldLength-((d^oldLength)&1))

		// Forward pass
		for k := L; k <= R; k += 2 {
			var x int
			if k == L || (k != R && m.vForward[k-1] < m.vForward[k+1]) {
				x = m.vForward[k+1]
			} else {
				x = m.vForward[k-1] + 1
			}
			y := x - k + newLength
			x += m.commonLenForward(oldStart+x, newStart+y, min(oldEnd-oldStart-x, newEnd-newStart-y))
			m.vForward[k] = x
		}

		if (oldLength-newLength)%2 != 0 {
			for k := L; k <= R; k += 2 {
				if oldLength-(d-1) <= k && k <= oldLength+(d-1) {
					if m.vForward[k]+m.vBackward[newLength+oldLength-k] >= oldLength {
						xx = m.vForward[k]
						kk = k
						td = 2*d - 1
						break outer
					}
				}
			}
		}

		// Backward pass
		for k := L; k <= R; k += 2 {
			var x int
			if k == L || (k != R && m.vBackward[k-1] < m.vBackward[k+1]) {
				x = m.vBackward[k+1]
			} else {
				x = m.vBackward[k-1] + 1
			}
			y := x - k + newLength
			x += m.commonLenBackward(oldEnd-1-x, newEnd-1-y, min(oldEnd-oldStart-x, newEnd-newStart-y))
			m.vBackward[k] = x
		}

		if (oldLength-newLength)%2 == 0 {
			for k := L; k <= R; k += 2 {
				if oldLength-d <= k && k <= oldLength+d {
					if m.vForward[oldLength+newLength-k]+m.vBackward[k] >= oldLength {
						xx = oldLength - m.vBackward[k]
						kk = oldLength + newLength - k
						td = 2 * d
						break outer
					}
				}
			}
		}
	}

	if td > 1 {
		yy := xx - kk + newLength
		oldDiff := (td + 1) / 2
		if xx > 0 && yy > 0 {
			if err := m.run(oldStart, oldStart+xx, newStart, newStart+yy, oldDiff, throwErr); err != nil {
				return err
			}
		}
		if oldStart+xx < oldEnd && newStart+yy < newEnd {
			if err := m.run(oldStart+xx, oldEnd, newStart+yy, newEnd, td-oldDiff, throwErr); err != nil {
				return err
			}
		}
	} else if td >= 0 {
		x, y := oldStart, newStart
		for x < oldEnd && y < newEnd {
			cl := m.commonLenForward(x, y, min(oldEnd-x, newEnd-y))
			if cl > 0 {
				m.addUnchanged(x, y, cl)
				x += cl
				y += cl
			} else if oldEnd-oldStart > newEnd-newStart {
				x++
			} else {
				y++
			}
		}
	} else {
		if throwErr {
			return ErrTooBig
		}
	}
	return nil
}

func (m *MyersLCS) addUnchanged(start1, start2, count int) {
	m.changes1.SetRange(m.start1+start1, m.start1+start1+count, false)
	m.changes2.SetRange(m.start2+start2, m.start2+start2+count, false)
}

func (m *MyersLCS) commonLenForward(oldIdx, newIdx, maxLen int) int {
	maxLen = min(maxLen, min(m.count1-oldIdx, m.count2-newIdx))
	x := oldIdx
	y := newIdx
	for x-oldIdx < maxLen && m.first[m.start1+x] == m.second[m.start2+y] {
		x++
		y++
	}
	return x - oldIdx
}

func (m *MyersLCS) commonLenBackward(oldIdx, newIdx, maxLen int) int {
	maxLen = min(maxLen, min(oldIdx+1, newIdx+1))
	x := oldIdx
	y := newIdx
	for oldIdx-x < maxLen && m.first[m.start1+x] == m.second[m.start2+y] {
		x--
		y--
	}
	return oldIdx - x
}

// Changes returns the change bit sets. Index 0 is for first, 1 is for second.
// A set bit means the element at that index was changed (not part of LCS).
func (m *MyersLCS) Changes() [2]*BitSet {
	return [2]*BitSet{m.changes1, m.changes2}
}
