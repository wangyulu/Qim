package apis

import (
	"fmt"
	"hash/crc32"
	"time"

	"github.com/kataras/iris/v12"
	"jinv/kim"
	"jinv/kim/logger"
	"jinv/kim/naming"
	"jinv/kim/services/router/conf"
	"jinv/kim/services/router/ipregion"
	"jinv/kim/wire"
)

const DefaultLocation = "中国"

type RouterApi struct {
	Naming naming.Naming

	IpRegion ipregion.IpRegion
	Config   conf.Router
}

type LookUpResp struct {
	UTC      int64    `json:"utc"`
	Location string   `json:"location"`
	Domains  []string `json:"domains"`
}

func (r *RouterApi) Lookup(c iris.Context) {
	token := c.Params().Get("token")
	ip := kim.RealIP(c.Request())

	var location conf.Country
	ipInfo, err := r.IpRegion.Search(ip)
	if err != nil || ipInfo.Country == "0" {
		location = DefaultLocation
	} else {
		location = conf.Country(ipInfo.Country)
	}

	regionId, ok := r.Config.Mapping[location]
	if !ok {
		c.StopWithError(iris.StatusForbidden, err)
		return
	}

	region, ok := r.Config.Regions[regionId]
	if !ok {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}

	idc := selectIdc(token, region)
	gateways, err := r.Naming.Find(wire.SNWGateway, fmt.Sprintf("IDC:%s", idc.ID))
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}

	hitGateways := selectGateways(token, gateways, 3)
	domains := make([]string, len(hitGateways))
	for i, gateway := range hitGateways {
		domains[i] = gateway.GetMeta()["domain"]
	}

	logger.WithFields(logger.Fields{
		"country":  location,
		"regionId": regionId,
		"idc":      idc.ID,
	})

	_, _ = c.JSON(LookUpResp{
		UTC:      time.Now().Unix(),
		Location: string(location),
		Domains:  domains,
	})
}

func selectIdc(token string, region *conf.Region) *conf.IDC {
	slot := hashcode(token) % len(region.Slots)

	i := region.Slots[slot]

	return &region.Idcs[i]
}

func selectGateways(token string, gateways []kim.ServiceRegistration, num int) []kim.ServiceRegistration {
	if len(gateways) <= num {
		return gateways
	}

	slots := make([]int, len(gateways)*10)
	for i := range gateways {
		for j := 0; j < 10; j++ {
			slots = append(slots, i)
		}
	}

	slot := hashcode(token) % len(slots)

	i := slots[slot]

	res := make([]kim.ServiceRegistration, 0, num)
	for len(res) < num {
		res = append(res, gateways[i])
		i++
		if i >= len(gateways) {
			i = 0
		}
	}

	return res
}

func hashcode(key string) int {
	hash := crc32.NewIEEE()

	_, _ = hash.Write([]byte(key))

	return int(hash.Sum32())
}
