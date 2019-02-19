package db

// ColumnValue - interface for values in columns
// Make sure you use the value in this interface instead of the pointer.
type ColumnValue interface {
	Compare(ColumnValue) int
}

type ColumnType struct {
	Name         string
	Index        int
	ZeroValue    ColumnValue
	Lines        int
	UniqueValues int
}

type DataTable struct {
	metadata []*ColumnType
	columns  []*Column
	names    map[string]int
}

func (dt *DataTable) AddColumn(ct *ColumnType) int {
	idx := len(dt.metadata)
	if dt.names == nil {
		dt.names = make(map[string]int)
	}
	dt.names[ct.Name] = idx
	dt.metadata = append(dt.metadata, ct)
	dt.columns = append(dt.columns, NewColumnZeroVal(ct.Lines, ct.UniqueValues, ct.ZeroValue))
	ct.Index = idx
	return idx
}

func (dt *DataTable) Insert(colindex int, id IDEntry, val ColumnValue, opts QueryOptions) IDEntry {
	col := dt.columns[colindex]
	col.Lock()
	if id == NewIDEntry {
		id = col.maxId + 1
	}
	col.SetVal(id, val, opts&INSERT_UPDATE != 0, opts&INSERT_ASYNC != 0)
	col.Unlock()
	return id
}

func (dt *DataTable) Select(colindex int, where ColumnValue, opts QueryOptions) IDIterator {
	col := dt.columns[colindex]
	de, ok := col.dict.In(where)
	if !ok {
		return nil
	}

	col.RLock()

	// TODO: lock in iterator

	// TODO: GT/GTE/LT/LTE iterators

	iter := col.IteratorWithFilterVal(DataEntry(de), opts&SELECT_DESC != 0, opts&SELECT_NEQ != 0)
	col.RUnlock()

	return iter
}

func (dt *DataTable) SelectN(colname string, where ColumnValue, opts QueryOptions) IDIterator {
	colidx, ok := dt.names[colname]
	if !ok {
		return nil
	}
	return dt.Select(colidx, where, opts)
}

func (dt *DataTable) Or(iters ...IDIterator) IDIterator {
	if len(iters) == 0 {
		return nil
	}
	return NewIteratorMerge(iters...)
}

func (dt *DataTable) And(iters ...IDIterator) IDIterator {
	if len(iters) == 0 {
		return nil
	}
	iter := NewIteratorIntersect(iters[0].Reversed())
	for _, it := range iters {
		iter.Append(it)
	}
	return iter
}

func (dt *DataTable) Sub(iter IDIterator, diffIters ...IDIterator) IDIterator {
	if iter == nil {
		return nil
	}

	var isec *IntersectIterator
	if it, ok := iter.(*IntersectIterator); ok {
		isec = it
	} else {
		isec = NewIteratorIntersect(iter.Reversed())
		isec.Append(iter)
	}

	for _, it := range diffIters {
		isec.AppendDiff(it)
	}
	return isec
}
