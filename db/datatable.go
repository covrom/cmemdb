package db

type ColumnValue interface {
	Eq(ColumnValue) bool
	Lt(ColumnValue) bool
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
