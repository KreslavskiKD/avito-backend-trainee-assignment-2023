package db_postgresql

import (
	pq "github.com/lib/pq"
)

var (
	NotFoundEror       = pq.ErrorCode("23503")
	AlreadyExistsError = pq.ErrorCode("23505")
)
