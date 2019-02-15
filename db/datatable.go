package db

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
