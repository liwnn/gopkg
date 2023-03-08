package sensitive

type mapChildren map[rune]*node

func (m mapChildren) insert(c rune) *node {
	n, ok := m[c]
	if ok {
		return n
	}
	n = &node{
		ch: c,
	}
	m[c] = n
	return n
}

func (m mapChildren) find(ch rune) *node {
	return m[ch]
}

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
}

func (n *node) insert(c rune) *node {
	index, found := n.children.find(c)
	if found {
		return n.children[index]
	} else {
		next := &node{ch: c, p: n}
		n.children.insertAt(index, next)
		return next
	}
}

func (n *node) find(ch rune) *node {
	index, found := n.children.find(ch)
	if found {
		return n.children[index]
	}
	return nil
}

type AhoCorasick struct {
	root mapChildren
}

func New() *AhoCorasick {
	return &AhoCorasick{root: make(mapChildren)}
}

func (ac *AhoCorasick) Add(word string) {
	var depth int
	var curNode *node
	for _, v := range word {
		c := ac.convert(v)
		if depth == 0 {
			curNode = ac.root.insert(c)
		} else {
			curNode = curNode.insert(c)
		}
		depth++
	}
	if curNode != nil {
		curNode.output = true
	}
}

func (ac *AhoCorasick) Build() {
	var queue = newTravq(64)
	for _, n := range ac.root {
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
			for t := n.p.fail; ; t = t.fail {
				if t == nil {
					if node := ac.root.find(n.ch); node != nil {
						n.fail = node
					}
					break
				} else {
					if node := t.find(n.ch); node != nil {
						n.fail = node
						break
					}
				}
			}
			if len(n.children) > 0 {
				queue.push(n.children)
			}
		}
	}
}

func (ac *AhoCorasick) Contains(text string) bool {
	var p *node
	for _, v := range text {
		if ac.isWhite(v) {
			continue
		}
		c := ac.convert(v)
		for {
			var node *node
			if p == nil {
				node = ac.root.find(c)
			} else {
				node = p.find(c)
			}

			if node != nil {
				p = node
				if p.output {
					return true
				}
				// 找到了，继续匹配下一个字符
				break
			}
			// 没找到
			// 1. 根节点，不用继续找了
			if p == nil {
				break
			}
			p = p.fail
			// 2. 存在失败节点
			if p != nil && p.output {
				return true
			}
		}
	}
	return false
}

func (ac *AhoCorasick) Replace(text string, ch rune) string {
	var runeText []rune
	var runeIndex = -1
	var p *node
	for _, v := range text {
		runeIndex++
		if ac.isWhite(v) {
			continue
		}
		c := ac.convert(v)
		for {
			var node *node
			if p == nil {
				node = ac.root.find(c)
			} else {
				node = p.find(c)
			}
			if node != nil {
				p = node
				if p.output {
					if runeText == nil {
						runeText = []rune(text)
					}
					for j := runeIndex; p != nil; j-- {
						for ; ac.isWhite(runeText[j]); j-- {
						}
						runeText[j] = ch
						p = p.p
					}
				}
				break
			}
			if p == nil {
				break
			}
			p = p.fail
			if p != nil && p.output {
				if runeText == nil {
					runeText = []rune(text)
				}
				for j := runeIndex - 1; p != nil; j-- {
					for ; ac.isWhite(runeText[j]); j-- {
					}
					runeText[j] = ch
					p = p.p
				}
			}
		}
	}
	if runeText != nil {
		return string(runeText)
	}
	return text
}

func (*AhoCorasick) isWhite(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

func (*AhoCorasick) convert(r rune) rune {
	// 全角-》半角
	if r == 12288 {
		return 32
	}
	if insideCode := r - 65248; insideCode >= 32 && insideCode <= 126 {
		return insideCode
	}

	// 转小写
	if r >= 'A' && r <= 'Z' {
		r += 'a' - 'A'
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
