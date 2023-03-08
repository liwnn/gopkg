package namesearch

import (
	"math/rand"
	"testing"
)

func sliceEqual(t *testing.T, a, b []uint64) {
	if len(a) == len(b) {
		for i := 0; i < len(a); i++ {
			if a[i] != b[i] {
				t.Fatalf("not equal")
			}
		}
	} else {
		t.Fatalf("not equal")
	}
}

func TestSearch(t *testing.T) {
	testNames := []string{
		"abc",
		"bcd",
		"bcde",
		"cb",
	}
	xx := New()
	for i, v := range testNames {
		xx.Add(uint64(i), v)
	}

	sliceEqual(t, xx.Search("bcd"), []uint64{1, 2})
	sliceEqual(t, xx.Search("a"), []uint64{0})
	sliceEqual(t, xx.Search("bc"), []uint64{0, 1, 2})
	sliceEqual(t, xx.Search("e"), []uint64{2})
	sliceEqual(t, xx.Search("cb"), []uint64{3})
	sliceEqual(t, xx.Search("ef"), []uint64{})
}

func TestRemove(t *testing.T) {
	xx := New()
	xx.Add(1, "aaa")
	xx.Add(2, "aa")
	sliceEqual(t, xx.Search("aa"), []uint64{1, 2})
	xx.Remove(1, "aaa")
	sliceEqual(t, xx.Search("aa"), []uint64{2})
}

func genTestName(N int) []string {
	names := make([]string, 0, N)
	for i := 0; i < N; i++ {
		var name = make([]rune, 7)
		for i := 0; i < len(name); i++ {
			n := rand.Intn('\u9fa5' - '\u4e00' + 1)
			name[i] = '\u9fa5' + rune(n)
		}
		names = append(names, string(name))
	}
	return names
}

var benchmarkNames = genTestName(20000)

func BenchmarkSearch(b *testing.B) {
	xx := New()
	for j, v := range benchmarkNames {
		xx.Add(uint64(j), v)
	}

	searchKey := string([]rune{'\u9fa5'})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		xx.Search(searchKey)
	}
}

func BenchmarkNewWithNames(b *testing.B) {
	ids := rand.Perm(len(benchmarkNames))
	name := make([]Name, 0, len(ids))
	for i, id := range ids {
		name = append(name, Name{
			ID:   uint64(id),
			Name: benchmarkNames[i],
		})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewWithNames(name)
	}
}

func BenchmarkAdd(b *testing.B) {
	ids := rand.Perm(len(benchmarkNames))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		xx := New()
		for j, v := range benchmarkNames {
			xx.Add(uint64(ids[j]), v)
		}
	}
}

func BenchmarkAddRemove(b *testing.B) {
	ids := rand.Perm(len(benchmarkNames))
	var nodes []*NameSearch
	for i := 0; i < b.N; i++ {
		xx := New()
		for j, v := range benchmarkNames {
			xx.Add(uint64(ids[j]), v)
		}
		nodes = append(nodes, xx)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j, v := range benchmarkNames {
			nodes[i].Remove(uint64(ids[j]), v)
		}
	}
}
