package ipregion

import (
	"github.com/lionsoul2014/ip2region/binding/golang/ip2region"
)

type IpRegion interface {
	Search(ip string) (*IpInfo, error)
}

type IpInfo struct {
	Country  string
	Region   string
	Province string
	City     string
	ISP      string
}

type Ip2region struct {
	region *ip2region.Ip2Region
}

func NewIp2region(path string) (IpRegion, error) {
	if path == "" {
		path = "ip2region.db"
	}

	region, err := ip2region.New(path)
	defer region.Close()
	if err != nil {
		return nil, err
	}

	return &Ip2region{
		region: region,
	}, nil
}

func (r *Ip2region) Search(ip string) (*IpInfo, error) {
	ipInfo, err := r.region.MemorySearch(ip)
	if err != nil {
		return nil, err
	}

	return &IpInfo{
		Country:  ipInfo.Country,
		Region:   ipInfo.Region,
		Province: ipInfo.Province,
		City:     ipInfo.City,
		ISP:      ipInfo.ISP,
	}, nil
}
