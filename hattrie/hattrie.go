package hattrie

import (
	"bytes"
	"unsafe"
)

const (
	// set the default number of slots in each container
	HASH_SLOTS            uint64 = 512
	_32_BYTES                    = 32
	_64_BYTES                    = 64
	triePackEntryCapacity        = 1512
	BUCKET_OVERHEAD              = 8
	BUCKET_SIZE                  = (HASH_SLOTS * 8) + BUCKET_OVERHEAD
	KEYS_IN_BUCKET               = 0
	BUCKET_SIZE_LIM              = 65536
)

func bitwiseHash(b []byte) uint32 {
	h := uint64(220373)
	for _, c := range b {
		h ^= (h << 5) + uint64(c) + (h >> 2)
	}
	return uint32((h & 0x7fffffff) & (HASH_SLOTS - 1))
}

type hashTable [HASH_SLOTS][]byte

func resizeArray(ht hashTable, idx, requiredIncrease uint32) {
	if ht[idx] == nil {
		if requiredIncrease <= _32_BYTES {
			ht[idx] = make([]byte, requiredIncrease, _32_BYTES)
		} else {
			numberOfBlocks := ((requiredIncrease - 1) >> 6) + 1
			ht[idx] = make([]byte, requiredIncrease, numberOfBlocks<<6)
		}
	} else {
		oldArraySize := uint32(len(ht[idx]))
		newArraySize := oldArraySize + requiredIncrease
		if oldArraySize <= _32_BYTES && newArraySize <= _64_BYTES && newArraySize > _32_BYTES {
			tmp := make([]byte, newArraySize, _64_BYTES)
			copy(tmp, ht[idx])
			ht[idx] = tmp
			return
		} else if newArraySize > _64_BYTES {
			numberOfBlocks := ((oldArraySize - 1) >> 6) + 1
			numberOfNewBlocks := ((newArraySize - 1) >> 6) + 1
			if numberOfNewBlocks > numberOfBlocks {
				tmp := make([]byte, newArraySize, numberOfNewBlocks<<6)
				copy(tmp, ht[idx])
				ht[idx] = tmp
			}
		}
	}
}

func hashLookup(ht hashTable, query []byte) bool {
	i := bitwiseHash(query)
	if ht[i] == nil {
		return false
	}
	array := ht[i]
	for len(array) > 1 && array[0] != 0 {
		// calculate the length of the current string in the array.
		// Up to the first two bytes can be used to store the length of the string
		ln := uint32(array[0])
		if ln >= 128 {
			ln = ((ln & 0x7f) << 8) | uint32(array[1])
			array = array[2:]
		} else {
			array = array[1:]
		}
		word := array[:ln]
		if bytes.Equal(word, query) {
			return true
		}
		array = array[ln:]
	}
	return false
}

func hashInsert(ht hashTable, query []byte) bool {
	// get the required slot.
	idx := bitwiseHash(query)
	if ht[idx] != nil {
		array := ht[idx]
		for len(array) > 1 && array[0] != 0 {
			// calculate the length of the current string in the array.
			// Up to the first two bytes can be used to store the length of the string
			ln := uint32(array[0])
			if ln >= 128 {
				ln = ((ln & 0x7f) << 8) | uint32(array[1])
				array = array[2:]
			} else {
				array = array[1:]
			}
			word := array[:ln]
			if bytes.Equal(word, query) {
				return false
			}
			array = array[ln:]
		}
	}
	lnq := uint32(len(query))
	lnadd := lnq + 1
	if lnq >= 128 {
		lnadd++
	}
	arroff := uint32(len(ht[idx]))
	resizeArray(ht, idx, lnadd)
	array := ht[idx]
	if lnq < 128 {
		array[arroff] = byte(lnq)
		array = array[1:]
	} else {
		array[arroff] = byte(lnq>>8) | 0x80
		array[arroff+1] = byte(lnq)
		array = array[2:]
	}
	copy(array, query)
	return true
}

type triePackEntry [256]*byte

type TriePack struct {
	array    [][]triePackEntry
	arrayIdx uint32
	counter  uint32
	rootTrie triePos
	numTries int
}

type triePos struct {
	i, j uint32
}

func (tp *TriePack) newTrie() triePos {
	cnt := tp.counter
	if cnt == triePackEntryCapacity {
		tp.arrayIdx++
		for tp.arrayIdx >= uint32(len(tp.array)) {
			tp.array = append(tp.array, nil)
		}
		tp.array[tp.arrayIdx] = make([]triePackEntry, triePackEntryCapacity)
		tp.counter = 0
	}
	tp.counter++
	return triePos{tp.arrayIdx, cnt}
}

func newBucket() []byte {
	return make([]byte, BUCKET_SIZE)
}

func full(b []byte) bool {
	_ = b[3]
	mPtr := *(*uintptr)(unsafe.Pointer(&b))
	consumed := *(*uint32)(unsafe.Pointer(mPtr))
	return consumed > BUCKET_SIZE_LIM

	// unsafe examples:
	// mPtr := *(*uintptr)(unsafe.Pointer(&b))
	// a0 := *(*int)(unsafe.Pointer(mPtr))
	// a1 := *(*int)(unsafe.Pointer(mPtr + 4))

	// d := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	// h := &reflect.SliceHeader{
	// 	Data: d.Data,
	// 	Len:  d.Len / 4,
	// 	Cap:  d.Cap / 4,
	// }
	// dd := *(*[]int32)(unsafe.Pointer(h))
}

func NewTrie() *TriePack {
	tp := &TriePack{
		array: [][]triePackEntry{
			make([]triePackEntry, triePackEntryCapacity),
		},
	}
	tp.rootTrie = tp.newTrie()
	tp.numTries = 1
	return tp
}
