package bitset

import (
	"math/rand"
	"testing"
)

func perm(n int) []uint {
	var l []uint
	for _, v := range rand.Perm(n) {
		l = append(l, uint(v))
	}
	return l
}

func TestBitSet(t *testing.T) {
	const size = 10000
	b := NewSize(size)

	for _, v := range perm(size) {
		b.Set(v)
	}
	for _, v := range perm(size) {
		if !b.Get(v) {
			t.Error("Set bit error")
		}
	}

	for _, v := range perm(size / 2) {
		b.Clear(v)
	}

	for i := 0; i < size/2; i++ {
		if b.Get(uint(i)) {
			t.Error("Clear bit error")
		}
	}

	for i := size / 2; i < size; i++ {
		if !b.Get(uint(i)) {
			t.Error("Clear bit error")
		}
	}
}

func TestClear(t *testing.T) {
	b := New()
	b.Set(8)
	b.Set(9)
	b.Clear(8)
	b.Clear(8)
	b.Clear(10)
	b.Clear(64)
	if b.Get(8) {
		t.Error("8 shoud be clear")
	}
	if !b.Get(9) {
		t.Error("9 should not be clear")
	}
}

func TestSize(t *testing.T) {
	b := New()
	b.Set(8)
	if b.Size() != unitBitsNum {
		t.Errorf("Size should be %v", unitBitsNum)
	}
	b.Clear(8)
	if b.Size() != 0 {
		t.Error("Size need clear to 0")
	}
	b.Set(63)
	if b.Size() != unitBitsNum {
		t.Errorf("Size should be %v", unitBitsNum)
	}
}

func TestLength(t *testing.T) {
	b := New()
	b.Set(8)
	if b.Length() != 9 {
		t.Error("Length")
	}
}

func TestCardinality(t *testing.T) {
	tot := uint(64*4 + 11)
	v := NewSize(tot)
	checkLast := true
	for i := uint(0); i < tot; i++ {
		sz := v.Cardinality()
		if sz != i {
			t.Errorf("Cardinality reported as %d, but it should be %d", sz, i)
			checkLast = false
			break
		}
		v.Set(i)
	}
	if checkLast {
		sz := v.Cardinality()
		if sz != tot {
			t.Errorf("After all bits set, size reported as %d, but it should be %d", sz, tot)
		}
	}
}

func TestReset(t *testing.T) {
	bs := New()
	bs.Set(8)
	bs.Reset()

	if bs.Get(8) {
		t.Error("Bit need clear 8")
	}
	if bs.Length() != 0 {
		t.Error("Lengh need clear")
	}
	if bs.Size() != 0 {
		t.Error("Size need clear")
	}
}

func TestNextClearBit(t *testing.T) {
	bs := New()
	bs.Set(1)
	bs.Set(2)

	if i := bs.NextClearBit(0); i != 0 {
		t.Errorf("NextClearBit(0) = %d, want 1", i)
	}

	if i := bs.NextClearBit(1); i != 3 {
		t.Errorf("NextClearBit(1) = %d, want 3", i)
	}

	if i := bs.NextClearBit(10000); i != 10000 {
		t.Errorf("NextClearBit(10000) = %d, want 10000", i)
	}

	for i := 0; i < unitBitsNum; i++ {
		bs.Set(uint(i))
	}
	if i := bs.NextClearBit(0); i != unitBitsNum {
		t.Errorf("NextClearBit(0) = %d, want %d", i, unitBitsNum)
	}

	bs.Set(unitBitsNum)

	if i := bs.NextClearBit(0); i != unitBitsNum+1 {
		t.Errorf("NextClearBit(0) = %d, want %d", i, unitBitsNum+1)
	}
}

func TestNextSetBit(t *testing.T) {
	bs := New()
	bs.Set(1)
	bs.Set(2)

	if i, ok := bs.NextSetBit(0); !ok || i != 1 {
		t.Errorf("NextSetBit(0) = %d, want 1", i)
	}

	if i, ok := bs.NextSetBit(1); !ok || i != 1 {
		t.Errorf("NextSetBit(1) = %d, want 1", 1)
	}

	if _, ok := bs.NextSetBit(3); ok {
		t.Errorf("NextSetBit(3) = %v, want false", ok)
	}

	if _, ok := bs.NextSetBit(10000); ok {
		t.Errorf("NextSetBit(10000) = %v, want false", ok)
	}

	for i := 0; i < unitBitsNum; i++ {
		bs.Clear(uint(i))
	}
	if _, ok := bs.NextSetBit(0); ok {
		t.Errorf("NextSetBit(0) = %v, want false", ok)
	}

	bs.Set(unitBitsNum)

	if i, ok := bs.NextSetBit(0); !ok || i != unitBitsNum {
		t.Errorf("NextSetBit(0) = %d, want %d", i, unitBitsNum)
	}
}

var N = 1000000

func newBitSet() *BitSet {
	s := NewSize(uint(N))
	for i := 0; i < len(s.values); i++ {
		for j := 0; j < 8; j++ {
			s.Set(uint(i<<unitByteSize + j))
		}
	}
	return s
}

func BenchmarkSet(b *testing.B) {
	s := newBitSet()
	n := perm(N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Set(n[i%N])
	}
}

func BenchmarkGet(b *testing.B) {
	s := newBitSet()
	n := perm(N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Get(n[i%N])
	}
}

func BenchmarkCardinality(b *testing.B) {
	s := newBitSet()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Cardinality()
	}
}

func BenchmarkClear(b *testing.B) {
	s := newBitSet()
	n := perm(N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Clear(n[i%N])
	}
}

func BenchmarkNextClearBit(b *testing.B) {
	s := newBitSet()
	n := perm(N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.NextClearBit(n[i%N])
	}
}

func BenchmarkNextSetBit(b *testing.B) {
	s := newBitSet()
	n := perm(N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.NextSetBit(n[i%N])
	}
}

func BenchmarkForeachSetBit(b *testing.B) {
	b.StopTimer()
	s := NewSize(10000)
	for i := 0; i < 10000; i += 3 {
		s.Set(uint(i))
	}
	b.StartTimer()
	for j := 0; j < b.N; j++ {
		s.ForeachSetBit(0, func(j uint) bool {
			return false
		})
	}
}
