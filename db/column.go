package db

import (
	"sync"
)

type DataEntry int32

const NullEntry DataEntry = -1

func remFunc(v, m uint32) (uint32, uint32) { return v % m, v / m } // l, h

func valFunc(l, h, m uint32) uint32 { return h*m + l }

type valEntry struct {
	rem uint32
	ids []IDEntry
}

type kvSet struct {
	id  IDEntry
	val DataEntry
}

type Column struct {
	sync.RWMutex

	// кластерный индекс, сортирован в порядке возрастания ключа (ID)
	// индекс коллекции - это ID
	// могут быть пропуски ID, в них DataEntry==empty
	cluster []DataEntry

	// индекс, по значению (DataEntry), отсортирован только в рамках одного bucket
	// все одинаковые значения находятся в одном bucket
	// позволяет быстро найти по значению все ID, отсортированные по возрастанию
	// индекс коллекции - значение DataEntry
	bucketsCount uint32
	values       [][]valEntry

	bmp   []uint64 // биткарта
	count []int32  // количества по idx=val

	dict *Dictonary

	useval bool
	use1b  bool // биткарта, 1 бит на значение
	use2b  bool // биткарта, 2 бит на значение
	use4b  bool // биткарта, 4 бит на значение

	minId IDEntry
	maxId IDEntry

	empty DataEntry // для use1b Contains работает просто как проверка границ, для остальных - проверяет на это пустое значение

	chset chan kvSet
}

func NewColumnZeroVal(lines, vals int, zeroval ColumnValue) *Column {
	dct := NewDictonary(vals)
	return NewColumnZeroDataEntry(lines, vals, dct, DataEntry(dct.Put(zeroval)))
}

func NewColumnZeroDataEntry(lines, vals int, dct *Dictonary, zeroval DataEntry) *Column {
	ret := &Column{
		minId: 0xffffffff,
		dict:  dct,
		empty: zeroval,
		chset: make(chan kvSet, 1000),
	}
	if vals <= 2 {
		ret.use1b = true
		ret.bmp = make([]uint64, 1+(lines>>6))
		ret.count = make([]int32, 2)
	} else if vals <= 4 {
		ret.use2b = true
		ret.bmp = make([]uint64, 1+(lines>>5))
		ret.count = make([]int32, 4)
	} else if vals <= 16 {
		ret.use4b = true
		ret.bmp = make([]uint64, 1+(lines>>4))
		ret.count = make([]int32, 16)
	} else {
		d := lines / vals // lines per one value
		switch {
		case d > 10000000:
			ret.bucketsCount = 1 << 20
		case d > 1000000:
			ret.bucketsCount = 1 << 16
		case d > 100000:
			ret.bucketsCount = 1 << 13
		case d > 10000:
			ret.bucketsCount = 1 << 10
		default:
			ret.bucketsCount = 1 << 8
		}
		ret.cluster = make([]DataEntry, 0, lines)
		ret.values = make([][]valEntry, ret.bucketsCount)
		ret.useval = true
	}
	go ret.workerSet()
	return ret
}

func (c *Column) workerSet() {
	for kv := range c.chset {
		c.set(kv.id, kv.val)
	}
}

func (c *Column) SetVal(id IDEntry, v ColumnValue, upd, async bool) {
	if v == nil {
		c.Set(id, c.empty, upd, async)
	} else {
		c.Set(id, DataEntry(c.dict.Put(v)), upd, async)
	}
}

func binSearchValEntryFirst(a []valEntry, x uint32) uint32 {
	n := uint32(len(a))
	i, j := uint32(0), n
	for i < j {
		h := (i + j) >> 1
		if a[h].rem < x {
			i = h + 1
		} else {
			j = h
		}
	}
	return i
}

func binSearchValEntryLast(a []valEntry, x uint32) uint32 {
	n := uint32(len(a))
	i, j := uint32(0), n
	for i < j {
		h := (i + j) >> 1
		if a[h].rem <= x {
			i = h + 1
		} else {
			j = h
		}
	}
	return i
}

func binApproxSearchIDEntry(a []IDEntry, x IDEntry) uint32 {
	n := uint32(len(a))
	if n == 0 {
		return 0
	}
	min, max := a[0], a[n-1]
	if x < min {
		return 0
	} else if x > max {
		return n
	}
	i, j := uint32(0), n
	if n > 96 {
		// интерполяционная проба границы, здесь x уже точно внутри границ
		offset := uint32(float64(n-1) * (float64(x-min) / float64(max-min)))
		probe := a[offset]
		if probe == x {
			return offset
		} else if probe < x {
			i = offset
			if (offset < 32) && (a[offset+32] > x) {
				j = offset + 32
			}
		} else {
			j = offset
			if (n-offset < 32) && (a[offset-32] < x) {
				i = offset - 32
			}
		}
	}
	for i < j {
		h := (i + j) >> 1
		if a[h] < x {
			i = h + 1
		} else {
			j = h
		}
	}
	return i
}

func (c *Column) setCluster(id IDEntry, v DataEntry, clearset bool) {
	for uint32(len(c.cluster)) <= uint32(id) {
		c.cluster = append(c.cluster, NullEntry)
	}
	c.cluster[uint32(id)] = v
}

// нельзя вызывать для бинарных!
func (c *Column) Delete(id IDEntry, oldv DataEntry) {
	bck, rem := remFunc(uint32(oldv), c.bucketsCount)
	cv := c.values[bck]
	ln := len(cv)
	ii := int(binSearchValEntryFirst(cv, rem))
	if ii < ln && cv[ii].rem == rem {
		lnids := len(cv[ii].ids)
		iids := int(binApproxSearchIDEntry(cv[ii].ids, id))
		if iids < lnids && cv[ii].ids[iids] == id {
			if iids < lnids-1 {
				copy(cv[ii].ids[iids:], cv[ii].ids[iids+1:])
			}
			cv[ii].ids = cv[ii].ids[:lnids-1]
		}
		c.values[bck] = cv
		c.setCluster(id, NullEntry, false)
	}
}

func (c *Column) set(id IDEntry, v DataEntry) {
	if c.maxId < id {
		c.maxId = id
	}

	if c.minId > id {
		c.minId = id
	}

	if c.use1b {
		pos, sub := id>>6, id&0x3f
		mask := uint64(1) << sub
		if v == 1 {
			c.bmp[pos] |= mask
		} else {
			c.bmp[pos] &^= mask
		}
		c.count[v]++
		return
	} else if c.use2b {
		pos, sub := id>>5, id&0x1f
		mask := uint64(3) << (sub * 2)
		c.bmp[pos] &^= mask
		c.bmp[pos] |= (uint64(v) & 0x3) << (sub * 2)
		c.count[v]++
		return
	} else if c.use4b {
		pos, sub := id>>4, id&0x0f
		mask := uint64(0x0f) << (sub * 4)
		c.bmp[pos] &^= mask
		c.bmp[pos] |= (uint64(v) & 0x0f) << (sub * 4)
		c.count[v]++
		return
	}

	bck, rem := remFunc(uint32(v), c.bucketsCount)
	cv := c.values[bck]
	ln := len(cv)
	ii := int(binSearchValEntryFirst(cv, rem))
	if ii < ln && cv[ii].rem == rem {
		// уже есть значение - пробуем добавить ID
		lnids := len(cv[ii].ids)
		iids := int(binApproxSearchIDEntry(cv[ii].ids, id))
		// если уже есть - не добавляем
		if !(iids < lnids && cv[ii].ids[iids] == id) {
			cv[ii].ids = append(cv[ii].ids, id)
			if iids < lnids {
				copy(cv[ii].ids[iids+1:], cv[ii].ids[iids:])
				cv[ii].ids[iids] = id
			}
		}
	} else {
		cv = append(cv, valEntry{
			rem: rem,
			ids: []IDEntry{id},
		})
		if ii < ln {
			copy(cv[ii+1:], cv[ii:])
			cv[ii] = valEntry{
				rem: rem,
				ids: []IDEntry{id},
			}
		}
	}
	c.values[bck] = cv

	c.setCluster(id, v, false)
}

func (c *Column) Set(id IDEntry, v DataEntry, upd, async bool) {
	if upd {
		if c.use1b || c.use2b || c.use4b {
			oldv := c.Get(id)
			c.count[oldv]--
		} else {
			oldv := c.Get(id)
			if oldv != NullEntry {
				if v == oldv {
					return
				}
				// удаляем oldv
				c.Delete(id, oldv)
			}
		}
	}

	if async {
		c.chset <- kvSet{id, v}
	} else {
		c.set(id, v)
	}
}

func (c *Column) Get(id IDEntry) DataEntry {
	switch {
	case c.useval:
		return c.cluster[uint32(id)]
	case c.use1b:
		pos, sub := id>>6, id&0x3f
		mask := uint64(1) << sub
		return DataEntry((c.bmp[pos] & mask) >> sub)
	case c.use2b:
		pos, sub := id>>5, id&0x1f
		mask := uint64(3) << (sub * 2)
		return DataEntry((c.bmp[pos] & mask) >> (sub * 2))
	case c.use4b:
		pos, sub := id>>4, id&0x0f
		mask := uint64(0x0f) << (sub * 4)
		return DataEntry((c.bmp[pos] & mask) >> (sub * 4))
	default:
		panic("unknown column for Get")
	}
}

func (c *Column) IsZero(v DataEntry) bool {
	return v == c.empty
}

func (c *Column) ZeroVal() DataEntry {
	return c.empty
}

func (c *Column) ToDictonary(s ColumnValue) DataEntry {
	return DataEntry(c.dict.Put(s))
}

func (c *Column) InDictonary(s ColumnValue) (DataEntry, bool) {
	i, ok := c.dict.In(s)
	return DataEntry(i), ok
}

func (c *Column) FromDictonary(idx DataEntry) ColumnValue {
	return c.dict.Get(DictIndex(idx))
}

func (c *Column) DictonaryCompare(x, y DataEntry) int {
	return c.dict.Compare(DictIndex(x), DictIndex(y))
}

func (c *Column) GetVal(id IDEntry) ColumnValue {
	de := c.Get(id)
	if de != NullEntry {
		return c.dict.Get(DictIndex(de))
	}
	return nil
}

func (c *Column) DictCardinality() int {
	return c.dict.Length()
}

func (c *Column) Contains(id IDEntry) bool {
	switch {
	case c.use1b:
		return int(id>>6) < len(c.bmp)
	case c.use2b:
		pos, sub := id>>5, id&0x1f
		mask := uint64(3) << (sub * 2)
		return DataEntry((c.bmp[pos]&mask)>>(sub*2)) != 0
	case c.use4b:
		pos, sub := id>>4, id&0x0f
		mask := uint64(0x0f) << (sub * 4)
		return DataEntry((c.bmp[pos]&mask)>>(sub*4)) != 0
	default:
		v := c.cluster[uint32(id)]
		return !(v == NullEntry || v == c.empty)
	}
}

func (c *Column) GetV(v DataEntry) []IDEntry {
	if c.use1b || c.use2b || c.use4b {
		panic("GetV is not defined for bitmap columns")
	}
	bck, rem := remFunc(uint32(v), c.bucketsCount)
	cv := c.values[bck]
	ln := len(cv)
	ii := int(binSearchValEntryFirst(cv, rem))
	if ii < ln && cv[ii].rem == rem {
		return cv[ii].ids
	}
	return nil
}

// FIXME: compare ColumnValues instead of DataEntry
// FIXME: monotonic fast values buckets by range of values
// TODO: buckets or t-tree?

func (c *Column) IterateVUp(v DataEntry, f func(v DataEntry, ids []IDEntry) bool) {
	if c.use1b || c.use2b || c.use4b {
		panic("IterateUp is not defined for bitmap columns")
	}
	bcnt := c.bucketsCount
	bck, rem := remFunc(uint32(v), c.bucketsCount)
	cv := c.values[bck]
	ln := uint32(len(cv))
	ii := binSearchValEntryFirst(cv, rem)
	for {
		for ii >= ln {
			bck++
			if bck >= uint32(len(c.values)) {
				break
			}
			cv = c.values[bck]
			ln = uint32(len(cv))
			ii = binSearchValEntryFirst(cv, rem)
		}
		if !f(DataEntry(valFunc(ii, cv[ii].rem, bcnt)), cv[ii].ids) {
			break
		}
		ii++
	}
}

func (c *Column) IterateVDown(v DataEntry, f func(v DataEntry, ids []IDEntry) bool) {
	if c.use1b || c.use2b || c.use4b {
		panic("IterateUp is not defined for bitmap columns")
	}
	bcnt := c.bucketsCount
	bck, rem := remFunc(uint32(v), c.bucketsCount)
	bcki := int32(bck)
	cv := c.values[bcki]
	ii := int32(binSearchValEntryLast(cv, rem))
	for {
		ii--
		for ii < 0 {
			bcki--
			if bcki < 0 {
				break
			}
			cv = c.values[bcki]
			ii = int32(binSearchValEntryLast(cv, rem)) - 1
		}
		if !f(DataEntry(valFunc(uint32(ii), cv[ii].rem, bcnt)), cv[ii].ids) {
			break
		}
	}
}

func (c *Column) GetCountV(v DataEntry) int32 {
	if c.use1b || c.use2b || c.use4b {
		return c.count[v]
	}
	return int32(len(c.GetV(v)))
}

func (c *Column) RangeVals(f func(v DataEntry, ids []IDEntry)) {
	if c.use1b || c.use2b || c.use4b {
		panic("RangeVals is not defined for bitmap columns")
	} else {
		bck := c.bucketsCount
		for i, bucket := range c.values {
			for _, val := range bucket {
				f(DataEntry(valFunc(uint32(i), val.rem, bck)), val.ids)
			}
		}
	}
}
