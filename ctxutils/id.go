package ctxutils

import (
	"strconv"
	"sync/atomic"
)

var id uint64

func NewId() string {
	return strconv.FormatUint(atomic.AddUint64(&id, 1), 10)
}
