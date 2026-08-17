// Copyright 2024 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package uio

import (
	"errors"
	"testing"
)

// TestBufferHas covers Has, including a negative length. Before the guard,
// len(data) >= n was always true for n < 0, so ReadN/Consume/CopyN went on to
// slice data[:n] with a negative bound and panicked.
func TestBufferHas(t *testing.T) {
	b := NewBuffer([]byte{1, 2, 3, 4})
	for _, tc := range []struct {
		n    int
		want bool
	}{
		{-1 << 30, false},
		{-8, false},
		{-1, false},
		{0, true},
		{4, true},
		{5, false},
	} {
		if got := b.Has(tc.n); got != tc.want {
			t.Errorf("Has(%d) = %v, want %v", tc.n, got, tc.want)
		}
	}
}

// TestBufferReadNNegative checks ReadN returns an error rather than panicking on
// a negative length.
func TestBufferReadNNegative(t *testing.T) {
	if _, err := NewBuffer([]byte{1, 2, 3, 4}).ReadN(-6); !errors.Is(err, ErrBufferTooShort) {
		t.Fatalf("ReadN(-6) err = %v, want %v", err, ErrBufferTooShort)
	}
}

// TestLexerConsumeNegative checks Consume and CopyN return nil and set an error
// rather than panicking on a negative length. This is the primitive under the
// nclient4 DHCP receive path, where a parsed length minus a header size can go
// negative on a malformed frame.
func TestLexerConsumeNegative(t *testing.T) {
	for _, n := range []int{-1, -6, -8} {
		l := NewBigEndianBuffer([]byte{1, 2, 3, 4})
		if v := l.Consume(n); v != nil {
			t.Errorf("Consume(%d) = %v, want nil", n, v)
		}
		if l.Error() == nil {
			t.Errorf("Consume(%d): Error() = nil, want an error", n)
		}

		if v := NewBigEndianBuffer([]byte{1, 2, 3, 4}).CopyN(n); v != nil {
			t.Errorf("CopyN(%d) = %v, want nil", n, v)
		}
	}
}

// TestBufferWriteNegative checks the write-side length guards. Before them,
// WriteN and Append sliced through make([]byte, n) and Preallocate through
// make with capacity n, so a negative length panicked.
func TestBufferWriteNegative(t *testing.T) {
	for _, n := range []int{-1, -8} {
		if v := NewBuffer(nil).WriteN(n); v != nil {
			t.Errorf("WriteN(%d) = %v, want nil", n, v)
		}
		if v := NewBigEndianBuffer(nil).Append(n); v != nil {
			t.Errorf("Append(%d) = %v, want nil", n, v)
		}
		NewBuffer(nil).Preallocate(n) // must not panic
	}
}
