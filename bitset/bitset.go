// Package bitset implements a vector of bits that grows as needed.
package bitset

import "math/bits"

const (
	unitByteSize = 6
	unitBitsNum  = 1 << unitByteSize
	unitBitsMask = unitBitsNum - 1
	unitMask     = 1<<64 - 1
)

// BitSet manages a compact array of bit values, which are represented as bool,
// where true indicates that the bit is on (1) and false indicates the bit is off (0).
type BitSet struct {
	values    []uint64
	onesCount uint
}

// New Creates a new bit set. All bits are false
func New() *BitSet {
	return NewSize(1)
}

// NewSize returns a new BitSet whose initial size has at least the specified size.
func NewSize(size uint) *BitSet {
	n := size >> unitByteSize
	if size&unitBitsMask != 0 {
		n++
	}
	return &BitSet{
		values: make([]uint64, n),
	}
}

// Set index to 1.
func (b *BitSet) Set(index uint) {
	unitIndex := int(index >> unitByteSize)
	if unitIndex >= len(b.values) {
		b.grow(unitIndex + 1)
	}
	x := uint64(1 << (index & unitBitsMask))
	if b.values[unitIndex]&x == 0 {
		b.values[unitIndex] |= x
		b.onesCount++
	}
}

func (b *BitSet) grow(size int) {
	if size <= cap(b.values) {
		b.values = b.values[:size]
	} else {
		v := make([]uint64, size)
		copy(v, b.values)
		b.values = v
	}
}

// Get true if index is set 1, or return false.
func (b *BitSet) Get(index uint) bool {
	unitIndex := index >> unitByteSize
	return unitIndex < uint(len(b.values)) && (b.values[unitIndex]&(1<<(index&unitBitsMask))) != 0
}

// Clear sets the bit specified by the index to 0.
func (b *BitSet) Clear(index uint) {
	unitIndex := index >> unitByteSize
	if unitIndex >= uint(len(b.values)) {
		return
	}
	x := b.values[unitIndex] & ^(1 << (index & unitBitsMask))
	if x == b.values[unitIndex] {
		return
	}
	b.values[unitIndex] = x
	b.onesCount--

	i := len(b.values) - 1
	for ; i >= 0 && b.values[i] == 0; i-- {
	}
	b.values = b.values[:i+1]
}

// Reset all bits to 0.
func (b *BitSet) Reset() {
	for i := range b.values {
		b.values[i] = 0
	}
	b.values = b.values[:0]
	b.onesCount = 0
}

// Cardinality returns the number of bits set to true.
func (b BitSet) Cardinality() uint {
	return b.onesCount
}

// Size return the number of bits of space actually in use by this BitSet.
func (b BitSet) Size() uint64 {
	return uint64(len(b.values)) << unitByteSize
}

// Length return the "logical size": the index of the highest set bit plus one.
func (b BitSet) Length() int {
	n := len(b.values)
	if n == 0 {
		return 0
	}
	return n<<unitByteSize - bits.LeadingZeros64(b.values[n-1])
}

// NextClearBit return the index of the first bit that is set to false that occurs on or after
// the specified starting index.
func (b BitSet) NextClearBit(fromIndex uint) uint {
	index := fromIndex >> unitByteSize
	valueLen := uint(len(b.values))
	if index >= valueLen {
		return fromIndex
	}
	v := b.values[index] | ((1 << (fromIndex & unitBitsMask)) - 1)
	for {
		if v != unitMask {
			return index<<unitByteSize +
				uint(bits.TrailingZeros64(^v)) // find the first bit that is set to 0
		}
		index++
		if index >= valueLen {
			return valueLen << unitByteSize
		}
		v = b.values[index]
	}
}

// NextSetBit returns the index of the first bit that is set to true that occurs on or after
// the specified starting index. If no such bit exists then false is returned.
// Use ForeachSetBit for traverse.
func (b BitSet) NextSetBit(fromIndex uint) (uint, bool) {
	index := fromIndex >> unitByteSize
	valueLen := uint(len(b.values))
	if index >= valueLen {
		return 0, false
	}
	v := b.values[index] & (unitMask << (fromIndex & unitBitsMask))
	for {
		if v != 0 {
			return index<<unitByteSize + uint(bits.TrailingZeros64(v)), true
		}
		index++
		if index >= valueLen {
			return 0, false
		}
		v = b.values[index]
	}
}

// ForeachSetBit calls the do function for each bit that is set to true.
// It is faster than use NextSetBit.
// param - do: return true to quit.
func (b BitSet) ForeachSetBit(fromIndex uint, do func(uint) bool) {
	index := fromIndex >> unitByteSize
	valueLen := uint(len(b.values))
	if index >= valueLen {
		return
	}
	v := b.values[index] & (unitMask << (fromIndex & unitBitsMask))
	for {
		if v != 0 {
			offset := index << unitByteSize
			for {
				if do(offset + uint(bits.TrailingZeros64(v))) { // if true break.
					return
				}
				v &= v - 1
				if v == 0 {
					break
				}
			}
		}
		index++
		if index >= valueLen {
			return
		}
		v = b.values[index]
	}
}
