package container

import (
	"jinv/kim"
	"jinv/kim/wire/pkt"
)

type HashSelector struct {
}

func (hash *HashSelector) Lookup(header *pkt.Header, srvs []kim.Service) string {
	ll := len(srvs)

	code := HashCode(header.ChannelId)

	return srvs[code%ll].ServiceID()
}
