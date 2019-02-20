package hattrie

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

type hashTable [][]byte

func resizeArray(ht hashTable, idx, arrayOffset, requiredIncrease uint32) hashTable {
	if arrayOffset == 0 {
		if requiredIncrease <= _32_BYTES {
			ht[idx] = make([]byte, _32_BYTES)
		} else {
			numberOfBlocks := ((requiredIncrease - 1) >> 6) + 1
			ht[idx] = make([]byte, numberOfBlocks<<6)
		}
	} else {
		oldArraySize := arrayOffset + 1
		newArraySize := arrayOffset + requiredIncrease
		// if the new array size can fit within the previously allocated 32-byte block,
		// then no memory needs to be allocated.
		if oldArraySize <= _32_BYTES && newArraySize <= _32_BYTES {
			return ht
		} else if oldArraySize <= _32_BYTES && newArraySize <= _64_BYTES {
			// if the new array size can fit within a 64-byte block, then allocate only a
			// single 64-byte block.
			tmp := make([]byte, _64_BYTES)
			copy(tmp, ht[idx][:oldArraySize])
			ht[idx] = tmp
			return ht
		} else if oldArraySize <= _64_BYTES && newArraySize <= _64_BYTES {
			// if the new array size can fit within a 64-byte block, then return
			return ht
		} else {
			// resize the current array by as many 64-byte blocks as required
			numberOfBlocks := ((oldArraySize - 1) >> 6) + 1
			numberOfNewBlocks := ((newArraySize - 1) >> 6) + 1
			if numberOfNewBlocks > numberOfBlocks {
				tmp := make([]byte, numberOfNewBlocks<<6)
				copy(tmp, ht[idx][:numberOfBlocks<<6])
				ht[idx] = tmp
			}
		}
	}
	return ht
}
