package container

import (
	"hash/crc32"

	"jinv/kim"
	"jinv/kim/wire/pkt"
)

type Selector interface {
	Lookup(*pkt.Header, []kim.Service) string
}

func HashCode(key string) int {
	hash32 := crc32.NewIEEE()

	// todo 这里也不需要处理error的情况？
	hash32.Write([]byte(key))

	return int(hash32.Sum32())
}
