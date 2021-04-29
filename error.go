package snowflake

import "errors"

var (
	ErrWorkerID       = errors.New("illegal worker id")
	ErrClockBackwards = errors.New("illegal clock")
	ErrBits           = errors.New("illegal bits")
)
