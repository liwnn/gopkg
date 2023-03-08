package darts

import (
	"github.com/liwnn/bitset"
)

const (
	MaxASCII = '\u007F'
)

type children []*node

func (c children) find(r rune) (index int, found bool) {
	i, j := 0, len(c)
	for i < j {
		h := int(uint(i+j) >> 1)
		c := c[h].ch
		if c <= r {
			i = h + 1
		} else {
			j = h
		}
	}

	if i > 0 && c[i-1].ch == r {
		return i - 1, true
	}
	return i, false
}

func (c *children) insertAt(index int, n *node) {
	*c = append(*c, nil)
	if index < len(*c) {
		copy((*c)[index+1:], (*c)[index:])
	}
	(*c)[index] = n
}

type node struct {
	ch       rune
	output   bool
	p        *node
	fail     *node
	children children
	state    uint32
}

func (n *node) insert(c rune) (*node, bool) {
	index, found := n.children.find(c)
	if found {
		return n.children[index], false
	} else {
		next := &node{ch: c, p: n}
		n.children.insertAt(index, next)
		return next, true
	}
}

func (n *node) find(ch rune) *node {
	index, found := n.children.find(ch)
	if found {
		return n.children[index]
	}
	return nil
}

type state struct {
	base  uint32
	check uint32
	fail  uint32
}

func (u *state) clear() {
	u.base = 1 << 30
}

func (u state) isUse() bool {
	return u.base&(1<<30) == 0
}

func (u *state) setOffset(offset uint32) {
	u.base &= 1 << 31
	u.base |= offset
}

func (u state) offset() uint32 {
	return u.base & (1<<31 - 1)
}

func (u *state) setLeaf() {
	u.base |= 1 << 31
}

func (u state) isLeaf() bool {
	return (u.base >> 31) == 1
}

func (u *state) setFail(fail uint32) {
	u.fail = fail
}

func (u state) getFail() uint32 {
	return u.fail
}

type DoubleArrayTrie struct {
	units []state

	used *bitset.BitSet
	root *node
}

func New() *DoubleArrayTrie {
	t := &DoubleArrayTrie{
		units: make([]state, 0xFFFF*4),
		used:  bitset.NewSize(0xFFFF * 4),
		root:  &node{},
	}
	for i := range t.units {
		t.units[i].clear()
	}
	return t
}

func (t *DoubleArrayTrie) AddWord(words ...string) {
	for _, word := range words {
		curNode := t.root
		for _, c := range word {
			curNode, _ = curNode.insert(t.convert(c))
		}
		if curNode != nil {
			curNode.output = true
		}
	}
}

func (t *DoubleArrayTrie) Build() {
	root := t.root
	t.used.Set(0)

	var level int
	var newLevel []*node
	var nextLevel = []*node{root}
	var nextK uint = 0
	for len(nextLevel) > 0 {
		newLevel = newLevel[:0]
		for _, v := range nextLevel {
			state := v.state
			var k uint32
			if level > 0 {
				n := t.used.NextClearBit(nextK)
				//last, k = t.findk2(n, v.children)
				k, nextK = t.findk(n, v.children)
			}
			t.units[state].setOffset(k)
			for _, n := range v.children {
				offset := (t.index(n.ch) + k) % uint32(len(t.units))
				t.units[offset].check = state
				n.state = offset
				t.used.Set(uint(offset))
				if n.output {
					t.units[offset].setLeaf()
				}
				if len(v.children) > 0 {
					newLevel = append(newLevel, n)
				}
			}
		}
		nextLevel, newLevel = newLevel, nextLevel
		level++
	}

	var maxLevelCount uint32 = 4096

	var queue = newTravq(maxLevelCount)
	for _, n := range root.children {
		n.fail = root
		if len(n.children) > 0 {
			queue.push(n.children)
		}
	}
	for {
		p := queue.pop()
		if p == nil {
			break
		}
		for _, n := range p {
			for q := n.p.fail; q != nil; q = q.fail {
				if node := q.find(n.ch); node != nil {
					n.fail = node
					t.units[n.state].setFail(node.state)
					break
				}
			}
			if len(n.children) > 0 {
				queue.push(n.children)
			}
		}
	}
}

func (t *DoubleArrayTrie) findk(begin uint, children children) (uint32, uint) {
	var cycle bool
	size := uint(len(t.units))
LOOP:
	for k := begin; !(cycle && k >= begin); {
		for _, n := range children {
			p := uint(t.index(n.ch))
			if p >= size {
				break LOOP
			}
			index := (p + k) % size
			if t.used.Get(index) {
				k = t.used.NextClearBit(k + 1)
				if k >= size {
					k = 0
					cycle = true
				}
				continue LOOP
			}
		}
		return uint32(k), k + 1
	}
	panic("not implemented")
}

func (t *DoubleArrayTrie) index(r rune) uint32 {
	return uint32(r)
}

func (t *DoubleArrayTrie) ContainsWord(word string) bool {
	var s uint32
	for _, c := range word {
		if t.isWhite(c) {
			continue
		}
		c = t.convert(c)
		for {
			base := t.units[s].offset()
			offset := (base + t.index(c)) % uint32(len(t.units))
			unit := t.units[offset]
			if unit.check == s {
				// found
				if unit.isLeaf() {
					return true
				}
				s = offset
				break
			}
			// not found
			if s == 0 { // root
				break
			}
			s = t.units[s].getFail()
			if t.units[s].isLeaf() {
				return true
			}
		}
	}
	return false
}

func (t *DoubleArrayTrie) ReplaceWord(text string, ch rune) string {
	var runeText []rune
	var runeIndex = -1
	var s uint32
	for _, c := range text {
		runeIndex++
		if t.isWhite(c) {
			continue
		}
		c = t.convert(c)
		for {
			base := t.units[s].offset()
			offset := (base + t.index(c)) % uint32(len(t.units))
			unit := t.units[offset]
			if unit.isUse() && unit.check == s {
				// found
				s = offset
				if unit.isLeaf() {
					if runeText == nil {
						runeText = []rune(text)
					}
					p := offset
					for j := runeIndex; p != 0; j-- {
						for ; t.isWhite(runeText[j]); j-- {
						}
						runeText[j] = ch
						p = t.units[p].check
					}
					s = 0
				}
				break
			}
			// not found
			if s == 0 { // root
				break
			}
			s = t.units[s].getFail()
			if t.units[s].isLeaf() {
				if runeText == nil {
					runeText = []rune(text)
				}
				p := s
				for j := runeIndex - 1; p != 0; j-- {
					for ; t.isWhite(runeText[j]); j-- {
					}
					runeText[j] = ch
					p = t.units[p].check
				}
			}
		}
	}
	if runeText != nil {
		return string(runeText)
	}
	return text
}

func (*DoubleArrayTrie) isWhite(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

func (*DoubleArrayTrie) convert(r rune) rune {
	// 全角-》半角
	if r == 12288 {
		return 32
	}
	if insideCode := r - 65248; insideCode >= 32 && insideCode <= 126 {
		return insideCode
	}

	// 转小写
	if r <= MaxASCII {
		if 'A' <= r && r <= 'Z' {
			r += 'a' - 'A'
		}
		return r
	}
	return r
}

type travq struct {
	buf               []children
	head, tail, count int
}

func newTravq(n uint32) *travq {
	if n > 0 && n&(n-1) != 0 {
		// 转为2的n次方
		n |= n >> 1
		n |= n >> 2
		n |= n >> 4
		n |= n >> 8
		n |= n >> 16
		n += 1
	}
	return &travq{buf: make([]children, n)}
}

func (q *travq) push(c children) {
	if q.count == len(q.buf) {
		q.resize()
	}
	q.buf[q.tail] = c
	q.tail = (q.tail + 1) & (len(q.buf) - 1)
	q.count++
}

func (q *travq) pop() children {
	if q.count == 0 {
		return nil
	}
	c := q.buf[q.head]
	q.buf[q.head] = nil
	q.head = (q.head + 1) & (len(q.buf) - 1)
	q.count--
	return c
}

func (q *travq) resize() {
	newbuf := make([]children, len(q.buf)<<1)
	if q.head < q.tail {
		copy(newbuf, q.buf[q.head:q.tail])
	} else {
		n := copy(newbuf, q.buf[q.head:])
		copy(newbuf[n:], q.buf[:q.tail])
	}
	q.head = 0
	q.tail = q.count
	q.buf = newbuf
}
