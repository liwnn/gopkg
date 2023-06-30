package bloomfilter

import (
	"math"

	"github.com/liwnn/gopkg/bitset"
)

// BloomFilter 布隆过滤器
type BloomFilter struct {
	bitSet    *bitset.BitSet // 位数组
	numHashes int            // hash函数个数
}

// New new
// @param n - 预估元素个数
// @param p - false positive(误判率)
func New(n uint64, p float64) *BloomFilter {
	if p <= 0 || p >= 1 {
		panic("The false positive rate must be in (0,1)")
	}
	ln2 := 0.693147180559945                                // ln2
	denom := 0.480453013918201                              // ln(2)^2
	m := math.Ceil(-1 * (float64(n) * math.Log(p)) / denom) // 位数组的位数
	k := math.Ceil(m / float64(n) * ln2)                    // hash函数个数
	return &BloomFilter{
		bitSet:    bitset.NewSize(uint(m)),
		numHashes: int(k),
	}
}

func (bf *BloomFilter) bloomHash(data []byte) (uint64, uint64) {
	return MurmurHash3_x64_128(data, 0)
}

// Add 增加元素
func (bf *BloomFilter) Add(key []byte) {
	h1, h2 := bf.bloomHash(key)
	for i := 0; i < bf.numHashes; i++ {
		// 双重散列法(Double Hashing): h(i,k) = (h1(k) + i*h2(k)) % TABLE_SIZE
		h := (h1 + uint64(i)*h2) % bf.bitSet.Size()
		bf.bitSet.Set(uint(h))
	}
}

// MayContain 是否有存在可能
func (bf *BloomFilter) MayContain(data []byte) bool {
	h1, h2 := bf.bloomHash(data)
	for i := 0; i < bf.numHashes; i++ {
		h := (h1 + uint64(i)*h2) % bf.bitSet.Size()
		if !bf.bitSet.Get(uint(h)) {
			return false
		}
	}
	return true
}
