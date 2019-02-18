package db

import (
	"sync"
)

type DictIndex uint32

type Dictonary struct {
	sync.RWMutex

	mm map[ColumnValue]DictIndex
	ms []ColumnValue
}

func NewDictonary(c int) *Dictonary {
	return &Dictonary{
		mm: make(map[ColumnValue]DictIndex, c),
		ms: make([]ColumnValue, 0, c),
	}
}

func (ld *Dictonary) Length() int {
	ld.RLock()
	l := len(ld.ms)
	ld.RUnlock()
	return l
}

func (ld *Dictonary) Put(b ColumnValue) DictIndex {
	ld.Lock()
	if i, ok := ld.mm[b]; ok {
		ld.Unlock()
		return i
	}
	i := len(ld.ms)
	ld.ms = append(ld.ms, b)
	ld.mm[b] = DictIndex(i)
	ld.Unlock()
	return DictIndex(i)
}

func (ld *Dictonary) In(b ColumnValue) (DictIndex, bool) {
	ld.RLock()
	if i, ok := ld.mm[b]; ok {
		ld.RUnlock()
		return i, true
	}
	ld.RUnlock()
	return 0, false
}

func (ld *Dictonary) Get(n DictIndex) ColumnValue {
	ld.RLock()
	if int(n) < len(ld.ms) {
		ld.RUnlock()
		return ld.ms[int(n)]
	}
	ld.RUnlock()
	return nil
}

func (ld *Dictonary) Compare(x, y DictIndex) int {
	return ld.Get(x).Compare(ld.Get(y))
}

func (ld *Dictonary) Delete(n DictIndex) {
	ld.Lock()
	ln := len(ld.ms)
	if int(n) < ln {
		b := ld.ms[n]
		if b != nil {
			ld.ms[n] = nil
			delete(ld.mm, b)
		}
	}
	ld.Unlock()
}
