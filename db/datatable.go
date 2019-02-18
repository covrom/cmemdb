package db

// ColumnValue - interface for values in columns
// Make sure you use the value in this interface instead of the pointer.
type ColumnValue interface {
	Compare(ColumnValue) int
}

type ColumnType struct {
	Name         string
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
	return idx
}
