package db

type QueryOptions uint8

const (
	INSERT_UPDATE QueryOptions = 1 << iota
	INSERT_ASYNC
	SELECT_DESC
	SELECT_NEQ
)
