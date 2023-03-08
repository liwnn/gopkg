package namesearch

import (
	"math/rand"
	"time"
)

const (
	defaultMaxLevel = 10   // (1/p)^MaxLevel >= maxNode
	defaultP        = 0.25 // SkipList P = 1/4

	posPerNode = 256
)

var rander = rand.New(rand.NewSource(time.Now().UnixNano()))

type node struct {
	positions []position
	forward   []*node
}

func (n *node) insert(id uint64, pos uint32) {
	j := len(n.positions)
	if j == 0 || id > n.positions[j-1].id {
		n.positions = append(n.positions, position{
			id:  id,
			pos: pos,
		})
		return
	}
	if id < n.positions[0].id {
		n.positions = n.positions[:j+1]
		copy(n.positions[1:], n.positions[:j])
		n.positions[0].id = id
		n.positions[0].pos = pos
	} else {
		index, found := n.find(id)
		if found {
			n.positions[index].addPos(pos)
			return
		}
		n.positions = n.positions[:j+1]
		copy(n.positions[index+1:], n.positions[index:j])
		n.positions[index].id = id
		n.positions[index].pos = pos
	}
}

func (n *node) insertIndex(index int, id uint64, pos uint32) {
	l := len(n.positions)
	n.positions = n.positions[:l+1]
	copy(n.positions[index+1:], n.positions[index:l])
	n.positions[index].id = id
	n.positions[index].pos = pos
}

func (n *node) delete(id uint64, pos uint32) {
	index, ok := n.find(id)
	if ok {
		n.positions[index].removePos(pos)
		if n.positions[index].pos == 0 {
			copy(n.positions[index:], n.positions[index+1:])
			n.positions = n.positions[:len(n.positions)-1]
		}
	}
}

func (n *node) find(id uint64) (int, bool) {
	i, j := 0, len(n.positions)
	for i < j {
		h := int(uint(i+j) >> 1)
		if n.positions[h].id <= id {
			i = h + 1
		} else {
			j = h
		}
	}

	if i > 0 && n.positions[i-1].id == id {
		return i - 1, true
	}
	return i, false
}

func (n *node) Len() int {
	return len(n.positions)
}

func (n *node) max() uint64 {
	return n.positions[len(n.positions)-1].id
}

func (n *node) min() uint64 {
	return n.positions[0].id
}

func (n *node) getPosition(index uint32) position {
	return n.positions[index]
}

type freeList struct {
	list  [16]*node
	count int
}

func (l *freeList) new(lvl int32, length, size int) *node {
	if l.count == 0 {
		return &node{
			forward:   make([]*node, lvl),
			positions: make([]position, length, size),
		}
	}
	index := l.count - 1
	n := l.list[index]
	l.list[index] = nil
	l.count--
	if int32(len(n.forward)) < lvl {
		n.forward = make([]*node, lvl)
	} else {
		n.forward = n.forward[:lvl]
	}
	n.positions = n.positions[:length:size]
	return n
}

func (l *freeList) free(node *node) {
	if l.count < cap(l.list) {
		for i := range node.forward {
			node.forward[i] = nil
		}
		node.positions = node.positions[:0]
		l.list[l.count] = node
		l.count++
	}
}

var defaultFreeList freeList

type skipData struct {
	header *node
	level  int32 // current max level
}

func newSkipData() *skipData {
	return &skipData{
		level: 1,
		header: &node{
			forward: make([]*node, defaultMaxLevel),
		},
	}
}

func newNode(lvl int32, length int) *node {
	return defaultFreeList.new(lvl, length, posPerNode)
}

func freeNode(n *node) {
	defaultFreeList.free(n)
}

func (sl *skipData) First() *node {
	return sl.header.forward[0]
}

func (sl *skipData) Search(key uint64) (*node, uint32) {
	x := sl.header
	for i := sl.level - 1; i >= 0; i-- {
		for y := x.forward[i]; y != nil && y.min() <= key; y = x.forward[i] {
			x = y
		}
	}

	if x != nil {
		index, ok := x.find(key)
		if ok {
			return x, uint32(index)
		}
	}
	return nil, 0
}

func (sl *skipData) Insert(id uint64, pos uint32) {
	var prev [defaultMaxLevel]*node
	x := sl.header
	for i := sl.level - 1; i >= 0; i-- {
		for y := x.forward[i]; y != nil && y.positions[len(y.positions)-1].id < id; y = y.forward[i] {
			x = y
		}
		prev[i] = x
	}

	y := x
	x = x.forward[0] // x.max >= id or x == nil
	if x != nil {
		if id == x.max() {
			x.positions[len(x.positions)-1].addPos(pos)
			return
		}
		// 优先插入左边的节点
		if id < x.min() {
			if y != sl.header && len(y.positions) < posPerNode {
				y.insert(id, pos)
				return
			}

			// x 还没满
			if x.Len() < posPerNode {
				x.insert(id, pos)
				return
			}
		} else {
			index, found := x.find(id)
			if found {
				x.positions[index].addPos(pos)
				return
			}
			// 没找到且x还没满
			if x.Len() < posPerNode {
				x.insertIndex(index, id, pos)
				return
			}

			// 没找到且x满了
			if index < posPerNode/2 {
				// 左半部分
				if y.Len() < posPerNode {
					// 左移一个节点
					y.positions = append(y.positions, x.positions[0])
					copy(x.positions, x.positions[1:index])
					x.positions[index].id = id
					x.positions[index].pos = pos
					return
				}
			} else {
				// 右半部分
				next := x.forward[0]
				if next != nil && next.Len() < posPerNode/2 {
					// 下一个节点小于一半
					offset := posPerNode - index
					nextOldLen := next.Len()
					next.positions = next.positions[:nextOldLen+offset]
					copy(next.positions[offset:], next.positions[:nextOldLen])
					copy(next.positions, x.positions[index:])
					x.positions = x.positions[:index+1]
					x.positions[index].id = id
					x.positions[index].pos = pos
					return
				}
			}

			// 分裂
			lvl := sl.randomLevel()
			if lvl > sl.level {
				for i := sl.level; i < lvl; i++ {
					prev[i] = sl.header
				}
				sl.level = lvl
			}

			rightCap := posPerNode - index
			n := newNode(lvl, rightCap)
			copy(n.positions, x.positions[index:])

			for i := int32(0); i < lvl; i++ {
				if prev[i].forward[i] == x {
					n.forward[i], x.forward[i] = x.forward[i], n
				} else {
					n.forward[i], prev[i].forward[i] = prev[i].forward[i], n
				}
			}

			x.positions = x.positions[:index+1]
			x.positions[index].id = id
			x.positions[index].pos = pos
			return
		}
	} else {
		if y != sl.header && y.Len() < posPerNode {
			y.insert(id, pos)
			return
		}
	}

	lvl := sl.randomLevel()
	if lvl > sl.level {
		for i := sl.level; i < lvl; i++ {
			prev[i] = sl.header
		}
		sl.level = lvl
	}

	x = newNode(lvl, 1)
	x.positions[0].id = id
	x.positions[0].pos = pos
	for i := int32(0); i < lvl; i++ {
		x.forward[i], prev[i].forward[i] = prev[i].forward[i], x
	}
}

func (sl *skipData) InsertSortPosition(nodes []position) {
	var id = nodes[0].id

	var prev [defaultMaxLevel]*node
	x := sl.header
	for i := sl.level - 1; i >= 0; i-- {
		for y := x.forward[i]; y != nil && y.max() < id; y = x.forward[i] {
			x = y
		}
		prev[i] = x
	}
	lvl := sl.randomLevel()
	if lvl > sl.level {
		for i := sl.level; i < lvl; i++ {
			prev[i] = sl.header
		}
		sl.level = lvl
	}
	x = &node{
		forward:   make([]*node, lvl),
		positions: nodes,
	}

	for i := int32(0); i < lvl; i++ {
		x.forward[i], prev[i].forward[i] = prev[i].forward[i], x
	}
}

func (sl *skipData) Delete(id uint64, pos uint32) bool {
	var prev [defaultMaxLevel]*node
	x := sl.header
	for i := sl.level - 1; i >= 0; i-- {
		for y := x.forward[i]; y != nil && y.max() < id; y = x.forward[i] {
			x = y
		}
		prev[i] = x
	}
	x = x.forward[0] // x == nil or x.max >= id
	if x != nil {
		x.delete(id, pos)
		if x.Len() == 0 {
			for i := int32(0); i < sl.level; i++ {
				if prev[i].forward[i] != x {
					break
				}
				prev[i].forward[i] = x.forward[i]
			}
			for sl.level > 1 && sl.header.forward[sl.level-1] == nil {
				sl.level--
			}
			freeNode(x)
		}
		return true
	}
	return false
}

func (sl *skipData) randomLevel() int32 {
	lvl := int32(1)
	for lvl < defaultMaxLevel && float32(rander.Uint32()&0xFFFF) < defaultP*0xFFFF {
		lvl++
	}
	return lvl
}
