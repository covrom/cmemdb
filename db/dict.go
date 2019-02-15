package db

import (
	"sync"
)

type Dictonary struct {
	sync.RWMutex
	mm map[ColumnValue]uint32
	ms []ColumnValue
}

func NewDictonary(c int) *Dictonary {
	return &Dictonary{
		mm: make(map[ColumnValue]uint32, c),
		ms: make([]ColumnValue, 0, c),
	}
}

func (ld *Dictonary) Length() int {
	ld.RLock()
	l := len(ld.ms)
	ld.RUnlock()
	return l
}

func (ld *Dictonary) Put(b ColumnValue) uint32 {
	ld.Lock()
	if i, ok := ld.mm[b]; ok {
		ld.Unlock()
		return i
	}
	i := len(ld.ms)
	ld.ms = append(ld.ms, b)
	ld.mm[b] = uint32(i)
	ld.Unlock()
	return uint32(i)
}

func (ld *Dictonary) In(b ColumnValue) (uint32, bool) {
	ld.RLock()
	if i, ok := ld.mm[b]; ok {
		ld.RUnlock()
		return i, true
	}
	ld.RUnlock()
	return 0, false
}

func (ld *Dictonary) Get(n uint32) ColumnValue {
	ld.RLock()
	if int(n) < len(ld.ms) {
		ld.RUnlock()
		return ld.ms[int(n)]
	}
	ld.RUnlock()
	return nil
}

func (ld *Dictonary) Compare(x, y uint32) int {
	return ld.Get(x).Compare(ld.Get(y))
}
