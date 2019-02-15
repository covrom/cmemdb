package db

type ElemHeapIDEntry struct {
	ID       IDEntry      // Value of this element.
	Iterator IDIterator // Which list this element comes from.
}

type IDEntryHeap struct {
	reverse bool
	Elems   []ElemHeapIDEntry
}

func NewIDEntryHeap(reverse bool, capacity int) *IDEntryHeap {
	return &IDEntryHeap{
		reverse: reverse,
		Elems:   make([]ElemHeapIDEntry, 0, capacity),
	}
}

func (h *IDEntryHeap) Clone() *IDEntryHeap {
	rv := NewIDEntryHeap(h.reverse, cap(h.Elems))
	for _, el := range h.Elems {
		v := ElemHeapIDEntry{
			ID:       el.ID,
			Iterator: el.Iterator.Clone(),
		}
		rv.Elems = append(rv.Elems, v)
	}
	return rv
}

func (h *IDEntryHeap) Len() int { return len(h.Elems) }
func (h *IDEntryHeap) Less(i, j int) bool {
	if h.reverse {
		return h.Elems[i].ID > h.Elems[j].ID
	} else {
		return h.Elems[i].ID < h.Elems[j].ID
	}
}
func (h *IDEntryHeap) Swap(i, j int) { h.Elems[i], h.Elems[j] = h.Elems[j], h.Elems[i] }
func (h *IDEntryHeap) Push(x ElemHeapIDEntry) {
	h.Elems = append(h.Elems, x)
}

func (h *IDEntryHeap) Pop() ElemHeapIDEntry {
	old := h.Elems
	n := len(old)
	x := old[n-1]
	h.Elems = old[0 : n-1]
	return x
}

func InitIDEntryHeap(h *IDEntryHeap) {
	n := h.Len()
	for i := n/2 - 1; i >= 0; i-- {
		downIDEntryHeap(h, i, n)
	}
}

func PushIDEntryHeap(h *IDEntryHeap, x ElemHeapIDEntry) {
	h.Push(x)
	upIDEntryHeap(h, h.Len()-1)
}

func PopIDEntryHeap(h *IDEntryHeap) ElemHeapIDEntry {
	n := h.Len() - 1
	h.Swap(0, n)
	downIDEntryHeap(h, 0, n)
	return h.Pop()
}

func RemoveIDEntryHeap(h *IDEntryHeap, i int) ElemHeapIDEntry {
	n := h.Len() - 1
	if n != i {
		h.Swap(i, n)
		if !downIDEntryHeap(h, i, n) {
			upIDEntryHeap(h, i)
		}
	}
	return h.Pop()
}

func FixIDEntryHeap(h *IDEntryHeap, i int) {
	if !downIDEntryHeap(h, i, h.Len()) {
		upIDEntryHeap(h, i)
	}
}

func upIDEntryHeap(h *IDEntryHeap, j int) {
	for {
		i := (j - 1) / 2 // parent
		if i == j || !h.Less(j, i) {
			break
		}
		h.Swap(i, j)
		j = i
	}
}

func downIDEntryHeap(h *IDEntryHeap, i0, n int) bool {
	i := i0
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && h.Less(j2, j1) {
			j = j2 // = 2*i + 2  // right child
		}
		if !h.Less(j, i) {
			break
		}
		h.Swap(i, j)
		i = j
	}
	return i > i0
}
