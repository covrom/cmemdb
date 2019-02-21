package hattrie

import (
	"bytes"
)

const (
	// set the default number of slots in each container
	HASH_SLOTS uint64 = 512
	_32_BYTES         = 32
	_64_BYTES         = 64
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

func newTrie() {

}
