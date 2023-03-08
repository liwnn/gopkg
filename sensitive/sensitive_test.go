package sensitive

import (
	"testing"
)

func TestSensitive(t *testing.T) {
	var ts = []struct {
		words    []string
		text     string
		result   string
		contains bool
	}{
		{
			[]string{"h", "she"}, "shs", "s*s", true,
		},
		{
			[]string{"h", "she"}, "p", "p", false,
		},
		{
			[]string{"h", "she"}, "h", "*", true,
		},
		{
			[]string{"her", "say", "she", "shr"}, "asherp", "a***rp", true,
		},
	}

	for _, v := range ts {
		s := New()
		for _, word := range v.words {
			s.Add(word)
		}
		s.Build()

		if s.Contains(v.text) != v.contains {
			t.Error("ContainsWord", v.text, "not", v.contains)
		}

		if result := s.Replace(v.text, '*'); result != v.result {
			t.Error("ReplaceWord", result, "!=", v.result)
		}
	}
}

func newAc() *AhoCorasick {
	s := New()
	for i := 'a'; i <= 'z'; i++ {
		s.Add(string([]rune{i}))
	}
	s.Build()
	return s
}

func BenchmarkBuild(b *testing.B) {
	bw := newAc()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bw.Build()
	}
}

var benchText = `ABCDEFabcdef`

func BenchmarkContains(b *testing.B) {
	bw := newAc()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bw.Contains(benchText)
	}
}

func BenchmarkReplace(b *testing.B) {
	bw := newAc()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bw.Replace(benchText, '*')
	}
}
