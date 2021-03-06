// Copyright 2019 Roman Covanyan https://github.com/covrom
// Use of this source code is governed by a MIT license
//
// Package hattrie implements HAT-trie, originally implemented in C
// by Dr. Nikolas Askitis https://github.com/naskitis
package hattrie

import (
	"bytes"
)

const (
	// set the default number of slots in each container
	HASH_SLOTS      uint64 = 512
	_32_BYTES              = 32
	_64_BYTES              = 64
	trieEntryCap           = 1512
	KEYS_IN_BUCKET         = 0
	BUCKET_SIZE_LIM        = 65536
	BUCKET_SIZE            = (HASH_SLOTS * 8)
)

func bitwiseHash(b []byte) uint32 {
	h := uint64(220373)
	for _, c := range b {
		h ^= (h << 5) + uint64(c) + (h >> 2)
	}
	return uint32((h & 0x7fffffff) & (HASH_SLOTS - 1))
}

type hashTable [][]byte

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

type flagTrie byte

const (
	FLAG_TRIE   flagTrie = 1
	FLAG_BUCKET flagTrie = 2
)

type triePackNode struct {
	ht     hashTable
	pos    triePos
	flag   flagTrie
	keycnt uint32
	eof    bool
}

type triePackEntry struct {
	nodes [256]triePackNode
	eof   bool
}

type TriePack struct {
	array     [][trieEntryCap]triePackEntry
	arrayIdx  uint32
	counter   uint32
	rootTrie  triePos
	numTries  int
	numBucket int
}

type triePos struct {
	i, j uint32
}

func (tp *TriePack) newTrie() triePos {
	cnt := tp.counter
	if cnt == trieEntryCap {
		tp.arrayIdx++
		for tp.arrayIdx >= uint32(len(tp.array)) {
			tp.array = append(tp.array, [trieEntryCap]triePackEntry{})
		}
		tp.counter = 0
	}
	tp.counter++
	return triePos{tp.arrayIdx, cnt}
}

func NewTrie() *TriePack {
	tp := &TriePack{
		array: [][trieEntryCap]triePackEntry{
			[trieEntryCap]triePackEntry{},
		},
	}
	tp.rootTrie = tp.newTrie()
	tp.numTries = 1
	return tp
}

func (tp *TriePack) search(word []byte) bool {
	cTrie := tp.rootTrie
	for i, ch := range word {
		// fetch the corresponding trie node pointer, if its null, then the string isn't in the HAT-trie.
		x := tp.array[cTrie.i][cTrie.j].nodes[ch]
		switch x.flag {
		case 0:
			return false
		case FLAG_TRIE:
			cTrie = x.pos
		case FLAG_BUCKET:
			// consume the lead character of the query string.
			if i == len(word) {
				return x.eof
			}
			return hashLookup(x.ht, word[i:])
		}
	}
	// if we have consumed the entire query string and haven't reached a container, then we must check the last trie node
	// we accessed to determine whether or not the string exists.
	return tp.array[cTrie.i][cTrie.j].eof
}

func (tp *TriePack) newContainer(cTrie triePos, path byte, word []byte) bool {
	x := triePackNode{flag: FLAG_BUCKET, ht: make(hashTable, BUCKET_SIZE)}
	if len(word) == 0 {
		x.eof = true
	} else {
		if hashInsert(x.ht, word) {
			x.keycnt++
		}
	}
	tp.array[cTrie.i][cTrie.j].nodes[path] = x
	tp.numBucket++
	return true
}

func (tp *TriePack) insert(word []byte) bool {
	cTrie := tp.rootTrie
	for i, ch := range word {
		x := tp.array[cTrie.i][cTrie.j].nodes[ch]
		switch x.flag {
		case 0:
			return tp.newContainer(cTrie, ch, word[i+1:])
		case FLAG_TRIE:
			cTrie = x.pos
		case FLAG_BUCKET:
			// if the string has been consumed, then we check the buckets end-of-string flag.
			// If it is set, then the string already exists, otherwise we set it to complete the insertion.
			if i == len(word) {
				if x.eof {
					return false
				} else {
					x.eof = true
					tp.array[cTrie.i][cTrie.j].nodes[ch] = x
					return true
				}
			}
			// attempt to insert what remains of the string into the containers hash table
			if hashInsert(x.ht, word[i+1:]) {
				x.keycnt++
				// see if we need to burst the container
				if x.keycnt > BUCKET_SIZE_LIM {
					x = tp.hatTrieBurst(ch, cTrie)
				}
				tp.array[cTrie.i][cTrie.j].nodes[ch] = x
				return true
			}
			// the string was found in the containers hash table
			return false
		}
	}
	// if we consumed the string prior to reaching any container, then we must set the end-of-string flag in the last
	// trie node accessed, unless its already set, in which case, the insertion fails.
	if tp.array[cTrie.i][cTrie.j].eof {
		return false
	} else {
		tp.array[cTrie.i][cTrie.j].eof = true
		return true
	}
}

func (tp *TriePack) hatTrieBurst(ch byte, cTrie triePos) triePackNode {
	nTrie := tp.newTrie()
	tp.numTries++
	x := triePackNode{flag: FLAG_TRIE, eof: tp.array[cTrie.i][cTrie.j].nodes[ch].eof}
	tp.array[cTrie.i][cTrie.j].nodes[ch].eof = false
	tp.array[cTrie.i][cTrie.j].nodes[ch].pos = nTrie
	tp.array[nTrie.i][nTrie.j].nodes[ch] = x
	// we are splitting a pure bucket/container
	tp.splitPure(cTrie, ch, nTrie)
	return tp.array[nTrie.i][nTrie.j].nodes[ch]
}

func (tp *TriePack) splitPure(cTrie triePos, ch byte, nTrie triePos) {
	array := tp.array[cTrie.i][cTrie.j].nodes[ch].ht
	
}
