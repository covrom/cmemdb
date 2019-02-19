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
