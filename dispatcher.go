package kim

import (
	"jinv/kim/wire/pkt"
)

type Dispatcher interface {
	Push(gateWay string, channels []string, p *pkt.LogicPkt) error
}
