package conf

import (
	"encoding/json"
	"io/ioutil"
)

type Zone struct {
	Id     string
	Weight int
}

type Route struct {
	RouteBy   string
	Zones     []Zone
	WhiteList map[string]string
	Slots     []int
}

func ReadRoute(path string) (*Route, error) {
	var conf struct {
		RouteBy   string
		Zones     []Zone
		WhiteList []struct {
			Key   string
			Value string
		}
	}

	bts, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bts, &conf)
	if err != nil {
		return nil, err
	}

	route := Route{
		RouteBy:   conf.RouteBy,
		Zones:     conf.Zones,
		WhiteList: make(map[string]string, len(conf.WhiteList)),
		Slots:     make([]int, 100),
	}

	for i, zone := range conf.Zones {
		for j := 0; j < zone.Weight; j++ {
			route.Slots = append(route.Slots, i)
		}
	}

	for _, l := range conf.WhiteList {
		route.WhiteList[l.Key] = l.Value
	}

	return &route, nil
}
