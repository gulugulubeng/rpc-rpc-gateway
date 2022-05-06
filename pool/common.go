package pool

import (
	"errors"
	"math"
)

// 默认值
const (
	// DefaultPoolSize 默认协程池容量
	DefaultPoolSize = math.MaxInt32
	// DefaultCleanIntervalTime 空闲协程默认清理时长
	DefaultCleanIntervalTime = 5
)

// Errors
var (
	ErrInvalidPoolSize   = errors.New("invalid size for pool")
	ErrInvalidPoolExpiry = errors.New("invalid expiry for pool")
	ErrInvalidPoolFunc   = errors.New("invalid func for pool")
	ErrPoolClosed        = errors.New("this pool has been closed")
)
