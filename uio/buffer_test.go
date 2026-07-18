// Copyright 2026 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package uio

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"testing"
)

// TestBufferHasBoundaries locks Has around its boundaries, including the
// negative guard and the zero-length case, so a later off-by-one (n > 0) or a
// dropped negative guard is caught.
func TestBufferHasBoundaries(t *testing.T) {
	b := NewBuffer([]byte{1, 2, 3})
	for _, tt := range []struct {
		n    int
		want bool
	}{
		{math.MinInt, false},
		{-1, false},
		{0, true},
		{1, true},
		{3, true},
		{4, false},
	} {
		if got := b.Has(tt.n); got != tt.want {
			t.Errorf("Has(%d) = %v, want %v", tt.n, got, tt.want)
		}
	}
}

// TestBufferNegativeLength checks that a negative length is rejected without
// disturbing the buffer. ReadN returns an ErrBufferTooShort error and leaves
// the data and length untouched, and Lexer.Consume records that error rather
// than reaching an invalid slice bound. Locking the no-side-effect behavior
// guards against a future "advance first, then error" partial consume.
func TestBufferNegativeLength(t *testing.T) {
	for _, n := range []int{-1, math.MinInt} {
		t.Run(fmt.Sprintf("ReadN(%d)", n), func(t *testing.T) {
			b := NewBuffer([]byte{1, 2, 3})
			before := bytes.Clone(b.Data())
			beforeLen := b.Len()

			got, err := b.ReadN(n)
			if got != nil {
				t.Errorf("ReadN(%d) = %v, want nil", n, got)
			}
			if !errors.Is(err, ErrBufferTooShort) {
				t.Errorf("ReadN(%d) error = %v, want ErrBufferTooShort", n, err)
			}
			if b.Len() != beforeLen {
				t.Errorf("ReadN(%d) changed Len from %d to %d", n, beforeLen, b.Len())
			}
			if !bytes.Equal(b.Data(), before) {
				t.Errorf("ReadN(%d) changed Data from %v to %v", n, before, b.Data())
			}
		})

		t.Run(fmt.Sprintf("Consume(%d)", n), func(t *testing.T) {
			l := NewBigEndianBuffer([]byte{1, 2, 3})
			if got := l.Consume(n); got != nil {
				t.Errorf("Consume(%d) = %v, want nil", n, got)
			}
			if l.Len() != 3 {
				t.Errorf("Consume(%d) changed buffer length to %d, want 3", n, l.Len())
			}
			if !errors.Is(l.Error(), ErrBufferTooShort) {
				t.Errorf("Consume(%d) error = %v, want ErrBufferTooShort", n, l.Error())
			}
		})
	}
}
