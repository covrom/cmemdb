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
	Metadata []*ColumnType
	Columns  []*Column
}

func (dt *DataTable) AddColumn(ct *ColumnType) {
	dt.Metadata = append(dt.Metadata, ct)
	dt.Columns = append(dt.Columns, NewColumnZeroVal(ct.Lines, ct.UniqueValues, ct.ZeroValue))
}
