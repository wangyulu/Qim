package serv

import (
	"hash/crc32"
	"math/rand"

	"jinv/kim"
	"jinv/kim/logger"
	"jinv/kim/services/gateway/conf"
	"jinv/kim/wire/pkt"
)

type RouteSelector struct {
	route *conf.Route
}

func NewRouteSelector(route *conf.Route) *RouteSelector {
	return &RouteSelector{
		route: route,
	}
}

func (s *RouteSelector) Lookup(header *pkt.Header, srvs []kim.Service) string {
	// 从header中取出Meta信息
	app, _ := pkt.FindMeta(header.Meta, MetaKeyApp)
	account, _ := pkt.FindMeta(header.Meta, MetaKeyAccount)
	if app == "" || account == "" {
		ri := rand.Intn(len(srvs))

		return srvs[ri].ServiceID()
	}

	log := logger.WithFields(logger.Fields{
		"app":     app,
		"account": account,
	})

	// 判断RouteBy是否命中白名单
	zone, ok := s.route.WhiteList[app.(string)]
	if !ok {
		var key string
		switch s.route.RouteBy {
		case MetaKeyApp:
			key = app.(string)
		case MetaKeyAccount:
			key = account.(string)
		default:
			key = account.(string)
		}

		// 通过权重计算出所属 Zone
		slot := hashcode(key) % len(s.route.Slots)

		i := s.route.Slots[slot]

		zone = s.route.Zones[i].Id

		log.Infoln("not hit a zone not in whileList", zone)
	} else {
		log.Infoln("hit a zone in whileList", zone)
	}

	// 过滤出所属 Zone 的所有 Servers
	zoneSrvs := filterSrvs(zone, srvs)
	if len(zoneSrvs) == 0 {
		noServerFoundErrorTotal.WithLabelValues(zone).Inc()

		log.Warnf("select a random service from all due to no service found in zone %s", zone)

		i := rand.Intn(len(srvs))

		return srvs[i].ServiceID()
	}

	// 从 ZoneSrvs 中选择一个服务
	srv := selectSrvs(account.(string), zoneSrvs)

	return srv.ServiceID()
}

func filterSrvs(zone string, srvs []kim.Service) []kim.Service {
	zoneSrvs := make([]kim.Service, 0)
	for _, srv := range srvs {
		if zone == srv.GetMeta()["zone"] {
			zoneSrvs = append(zoneSrvs, srv)
		}
	}

	return zoneSrvs
}

func selectSrvs(account string, srvs []kim.Service) kim.Service {
	slots := make([]int, 0, len(srvs)*10)
	for i := range srvs {
		for j := 0; j < 10; j++ {
			slots = append(slots, i)
		}
	}

	slot := hashcode(account) % len(slots)

	return srvs[slots[slot]]
}

func hashcode(key string) int {
	hash := crc32.NewIEEE()

	_, _ = hash.Write([]byte(key))

	return int(hash.Sum32())
}
