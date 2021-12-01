package conf

import (
	"encoding/json"
	"io/ioutil"
)

type IDC struct {
	ID     string
	Weight int
}

type Region struct {
	ID    string
	Idcs  []IDC
	Slots []byte
}

type Country string

type Mapping struct {
	Region    string
	Locations []string
}

type Router struct {
	Mapping map[Country]string
	Regions map[string]*Region
}

func LoadMapping(path string) (map[Country]string, error) {
	bts, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var mappings []*Mapping
	err = json.Unmarshal(bts, &mappings)
	if err != nil {
		return nil, err
	}

	mapping := make(map[Country]string)
	for _, val := range mappings {
		for _, loc := range val.Locations {
			mapping[Country(loc)] = val.Region
		}
	}

	return mapping, nil
}

func LoadRegions(path string) (map[string]*Region, error) {
	bts, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var regions []*Region
	err = json.Unmarshal(bts, &regions)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*Region)
	for _, region := range regions {
		res[region.ID] = region
		for i, idc := range region.Idcs {
			shard := make([]byte, idc.Weight)
			for j := 0; j < idc.Weight; j++ {
				shard[j] = byte(i)
			}

			region.Slots = append(region.Slots, shard...)
		}
	}

	return res, nil
}
