package namesearch

import (
	"sort"
)

const (
	MaxNameLen = 32

	bucketSize = 0xFFFF
)

type position struct {
	id  uint64
	pos uint32
}

func (n *position) addPos(pos uint32) {
	n.pos |= pos
}

func (n *position) removePos(pos uint32) {
	n.pos &= ^pos
}

type searchPos struct {
	position

	node  *node
	index uint32
}

func (s *searchPos) init(node *node) {
	s.node = node
	s.index = 0
	s.position = node.getPosition(0)
}

func (s *searchPos) getID() uint64 {
	return s.id
}

func (s *searchPos) getPos() uint32 {
	return s.pos
}

func (s *searchPos) next() (uint64, bool) {
	s.index++
	if s.index >= uint32(s.node.Len()) {
		s.node = s.node.forward[0]
		if s.node == nil {
			return 0, false
		}
		s.index = 0
	}

	s.position = s.node.getPosition(s.index)
	return s.id, true
}

type NameSearch struct {
	bucket   [bucketSize]*skipData
	overflow map[rune]*skipData

	result   [4096]uint64
	preAlloc [MaxNameLen]searchPos
}

func New() *NameSearch {
	return &NameSearch{}
}

type Name struct {
	ID   uint64
	Name string
}

func NewWithNames(names []Name) *NameSearch {
	sort.Slice(names, func(i, j int) bool {
		return names[i].ID < names[j].ID
	})

	var bucket [bucketSize][]position
	var overflow = make(map[rune][]position)
	var batchNum = posPerNode
	xx := New()
	for _, name := range names {
		var pos uint32 = 1
		for _, r := range name.Name {
			if r < bucketSize {
				if len(bucket[r]) > 0 && bucket[r][len(bucket[r])-1].id == name.ID {
					bucket[r][len(bucket[r])-1].addPos(pos)
					continue
				}
				if len(bucket[r]) >= batchNum {
					xx.getOrNewSortList(r).InsertSortPosition(bucket[r])
					bucket[r] = make([]position, 0, batchNum)
				} else if len(bucket[r]) == 0 {
					bucket[r] = make([]position, 0, batchNum)
				}
				bucket[r] = append(bucket[r], position{
					id:  name.ID,
					pos: pos,
				})
			} else {
				overflow[r] = append(overflow[r], position{
					id:  name.ID,
					pos: pos,
				})
			}
			pos <<= 1
		}
	}
	for k, v := range bucket {
		if len(v) > 0 {
			xx.getOrNewSortList(rune(k)).InsertSortPosition(v)
			bucket[k] = nil
		}
	}
	for k, v := range overflow {
		if len(v) > 0 {
			xx.getOrNewSortList(k).InsertSortPosition(v)
			overflow[k] = nil
		}
	}
	return xx
}

func (ac *NameSearch) Add(id uint64, name string) {
	var i int
	for _, r := range name {
		l := ac.getSortList(r)
		if l == nil {
			l = ac.newSortList(r)
		}

		l.Insert(id, 1<<i)
		i++
	}
}

// 取有序链表的交集，需要保存结构的话，要拷贝结果
func (ac *NameSearch) Search(name string) []uint64 {
	// 1. 找出所有链表开始id的最大值
	var maxID uint64 // 所有链表head最大的id
	var i int
	for _, v := range name {
		l := ac.getSortList(v)
		if l == nil { // 某个链表为空，说明交集肯定是空
			return nil
		}
		ac.preAlloc[i].init(l.First())
		if i == 0 {
			maxID = ac.preAlloc[i].getID()
		} else {
			id := ac.preAlloc[i].getID()
			for id < maxID {
				var ok bool
				id, ok = ac.preAlloc[i].next()
				if !ok {
					return nil
				}
			}
			if id > maxID {
				maxID = id
			}
		}
		i++
	}
	x := ac.preAlloc[:i]

	ids := ac.result[:0]
	for {
		// 2. 所有链表都对齐id相同
		for {
			oldMaxID := maxID
			for i := 0; i < len(x); i++ {
				id := x[i].getID()
				for id < maxID {
					var ok bool
					id, ok = x[i].next()
					if !ok { // 某个链表为空，说明交集没有了
						return ids
					}
				}
				if id > maxID {
					maxID = id
				}
			}
			if maxID == oldMaxID {
				break
			}
		}

		// 3. 计算结果，对于同样的id，如果pos是连续的，则满足
		if len(x)%2 == 0 {
			lastPos := x[0].getPos() << 1 & x[1].getPos()
			for j := 2; j < len(x); j += 2 {
				lastPos = lastPos << 1 & x[j].getPos()
				if lastPos == 0 { // 不连续则失败
					break
				}
				lastPos = lastPos << 1 & x[j+1].getPos()
				if lastPos == 0 { // 不连续则失败
					break
				}
			}
			if lastPos != 0 { // 成功
				ids = append(ids, x[0].getID())
			}
		} else {
			lastPos := x[0].getPos()
			for j := 1; j < len(x); j += 2 {
				lastPos = lastPos << 1 & x[j].getPos()
				if lastPos == 0 { // 不连续则失败
					break
				}
				lastPos = lastPos << 1 & x[j+1].getPos()
				if lastPos == 0 { // 不连续则失败
					break
				}
			}
			if lastPos != 0 { // 成功
				ids = append(ids, x[0].getID())
			}
		}

		maxID++
	}
}

func (ac *NameSearch) Remove(id uint64, name string) bool {
	var preAlloc [MaxNameLen]*skipData
	// 1. 找出所有链表开始id的最大值
	var i int
	for _, v := range name {
		l := ac.getSortList(v)
		if l == nil { // 某个链表为空，说明交集肯定是空
			return false
		}
		n, index := l.Search(id)
		if n == nil {
			return false
		}
		if n.getPosition(index).pos&(1<<i) == 0 {
			return false
		}
		preAlloc[i] = l
		i++
	}

	// 2. 删除对应位置
	var x = preAlloc[:i]
	for i, l := range x {
		l.Delete(id, 1<<i)
	}
	return true
}

func (ac *NameSearch) newSortList(r rune) *skipData {
	l := newSkipData()
	if r < bucketSize {
		ac.bucket[r] = l
	} else {
		if ac.overflow == nil {
			ac.overflow = make(map[rune]*skipData)
		}
		ac.overflow[r] = l
	}
	return l
}

func (ac *NameSearch) getSortList(r rune) *skipData {
	if r < bucketSize {
		return ac.bucket[r]
	}
	return ac.overflow[r]
}

func (ac *NameSearch) getOrNewSortList(r rune) *skipData {
	l := ac.getSortList(r)
	if l == nil {
		l = ac.newSortList(r)
	}
	return l
}

func (ac *NameSearch) deleteSortList(r rune) {
	if r < bucketSize {
		ac.bucket[r] = nil
	} else {
		delete(ac.overflow, r)
	}
}
