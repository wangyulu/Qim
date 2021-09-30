package wire

import (
	"math"
	"sync/atomic"
)

type sequence struct {
	num uint32
}

func (s *sequence) Next() uint32 {
	next := atomic.AddUint32(&s.num, 1)

	if next == math.MaxUint32 {
		// todo 这里再一次加强验证的目的是什么
		if atomic.CompareAndSwapUint32(&s.num, next, 1) {
			return 1
		}

		return s.Next()
	}

	return next
}

var Seq = sequence{1}
