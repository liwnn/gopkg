package darts

import (
	"bufio"
	"os"
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
			s.AddWord(word)
		}
		s.Build()

		if s.ContainsWord(v.text) != v.contains {
			t.Error("ContainsWord", v.text, "not", v.contains)
		}

		if result := s.ReplaceWord(v.text, '*'); result != v.result {
			t.Error("ReplaceWord", result, "!=", v.result)
		}
	}
}

func newAc() *DoubleArrayTrie {
	f, err := os.Open("../dict.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	s := New()
	r := bufio.NewReader(f)
	for {
		line, _, err := r.ReadLine()
		if err != nil {
			break
		}
		s.AddWord(string(line))
	}
	s.Build()
	return s
}

func BenchmarkBuild(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newAc()
	}
}

var benchText = `在计算机科学中，Aho–Corasick算法是由Alfred V. Aho和Margaret J.Corasick 发明的
字符串搜索算法，用于在输入的一串字符串中匹配有限组“字典”中的子串 [1]  。它与普通字符串匹配的不同点在于
同时与所有字典串进行匹配。算法均摊情况下具有近似于线性的时间复杂度，约为字符串的长度加所有匹配的数量。
然而由于需要找到所有匹配数，如果每个子串互相匹配（如字典为a，aa，aaa，aaaa，输入的字符串为aaaa），
算法的时间复杂度会近似于匹配的二次函数。`

func BenchmarkContains(b *testing.B) {
	bw := newAc()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bw.ContainsWord(benchText)
	}
}

func BenchmarkReplace(b *testing.B) {
	bw := newAc()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bw.ReplaceWord(benchText, '*')
	}
}
