package bloomfilter

import (
	"testing"
)

func TestBloomFilter(t *testing.T) {
	bf := New(1000, 0.01)
	n1 := []byte("Hurst")
	n2 := []byte("Peek")
	n3 := []byte("Beaty")
	bf.Add(n1)
	bf.Add(n3)
	n1b := bf.MayContain(n1)
	n2b := bf.MayContain(n2)
	n3b := bf.MayContain(n3)
	if !n1b {
		t.Errorf("%s should be in.", n1)
	}
	if n2b {
		t.Errorf("%s should not be in.", n2)
	}
	if !n3b {
		t.Errorf("%s should be in the second time we look.", n3)
	}
}
