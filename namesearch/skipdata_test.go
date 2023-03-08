package namesearch

import "testing"

func testSkipDataNodeNum(l *skipData) int {
	var count int
	for n := l.First(); n != nil; n = n.forward[0] {
		count++
	}
	return count
}

func TestSkipData_Insert(t *testing.T) {
	l := newSkipData()
	for i := 0; i < posPerNode; i++ {
		l.Insert(uint64(i), 1)
	}
	if testSkipDataNodeNum(l) != 1 {
		t.Error("Insert")
	}

	for i := posPerNode + 2; i < posPerNode*2+2; i++ {
		l.Insert(uint64(i), 1)
	}
	if testSkipDataNodeNum(l) != 2 {
		t.Error("Insert")
	}

	l.Insert(posPerNode, 1)
	if testSkipDataNodeNum(l) != 3 {
		t.Error("Insert")
	}

	l.Insert(posPerNode+1, 1)
	if testSkipDataNodeNum(l) != 3 {
		t.Error("Insert")
	}

	l.Insert(posPerNode+5, 1)
	if testSkipDataNodeNum(l) != 3 {
		t.Error("Insert")
	}

	n, i := l.Search(posPerNode)
	var result position
	result.id = posPerNode
	result.pos = 1
	if n == nil || n.positions[i] != result {
		panic("xxx")
	}

	l.Delete(posPerNode, 1)
	l.Delete(posPerNode+1, 1)
	if testSkipDataNodeNum(l) != 2 {
		t.Error("Insert")
	}

	l.Delete(posPerNode+5, 1)
	l.Insert(posPerNode*2+3, 1)
	if testSkipDataNodeNum(l) != 2 {
		t.Error("Insert")
	}

	l.Insert(posPerNode+5, 1)
	if testSkipDataNodeNum(l) != 3 {
		t.Error("Insert")
	}
}

func TestSkipData_Delete(t *testing.T) {
	l := newSkipData()
	l.Insert(2, 1)
	l.Insert(3, 1)
	l.Insert(1, 1)
	l.Insert(4, 1)
	l.Insert(5, 1)
	l.Insert(6, 1)

	l.Delete(2, 1)
	l.Delete(4, 1)
	l.Delete(6, 1)
	l.Delete(1, 1)
	l.Delete(5, 1)
	l.Delete(3, 1)

	if testSkipDataNodeNum(l) != 0 {
		t.Error("delete failed")
	}
}
